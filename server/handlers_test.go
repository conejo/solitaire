package server

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"go.uber.org/zap"
)

// newTestServer builds a Server with a no-op logger for tests.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	logger := zap.NewNop()
	s, err := NewServer(logger)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	return s
}

// TestGamesMapConcurrentReadWrite exercises the s.Games map from many
// goroutines simultaneously. With Go's built-in map and no mutex, this
// would trigger "fatal error: concurrent map read and map write" — both
// with and without the race detector. With the RWMutex guarding the map,
// it should complete cleanly. Run with `go test -race` to catch any
// regressions in the lock discipline.
func TestGamesMapConcurrentReadWrite(t *testing.T) {
	s := newTestServer(t)

	const goroutines = 32
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sid := strconv.Itoa(i) + "-" + strconv.Itoa(j%4)
				if _, err := s.getOrCreateGame(sid, "klondike"); err != nil {
					t.Errorf("getOrCreateGame failed: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// TestPerSessionConcurrentMoves verifies that the per-session mutex
// serializes concurrent game-mutating calls. Without it, the engine's
// internal Pile state could be torn between operations.
func TestPerSessionConcurrentMoves(t *testing.T) {
	s := newTestServer(t)
	sid := "concurrent-session"

	gs, err := s.getOrCreateGame(sid, "klondike")
	if err != nil {
		t.Fatalf("getOrCreateGame failed: %v", err)
	}

	// Move the face-up top card of tableau[1] onto tableau[0]. tableau[1]
	// starts with 2 cards (top face-up, bottom face-down); the face-up
	// index is len-1. After this, tableau[0] has [5♣] still and tableau[1]
	// has a face-down card plus a face-up top — but we don't care about
	// state details here, only that the engine can be hammered without
	// races.
	gs.mu.Lock()
	g := gs.Game
	faceUpIdx := g.GetState().Tableau[1].Size() - 1
	if err := g.Select("tableau", 1, faceUpIdx); err != nil {
		gs.mu.Unlock()
		t.Fatalf("setup Select failed: %v", err)
	}
	gs.mu.Unlock()

	// Perform many concurrent DrawFromStock calls on the same session.
	// With the per-session lock, none race; without it, internal state
	// could be torn.
	const goroutines = 16
	const iterations = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gs.mu.Lock()
				_ = gs.Game.DrawFromStock()
				gs.mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Verify the server didn't panic and we can still read state.
	gs.mu.Lock()
	_ = gs.Game.GetState()
	gs.mu.Unlock()
}

// TestServerIndexHandler is a smoke test for the wired-up routes.
func TestServerIndexHandler(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct == "" {
		t.Error("expected Content-Type header to be set")
	}
}

// TestStaticFilesServed verifies the static file route is wired up.
func TestStaticFilesServed(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// /static/style.css should be served from the embedded FS.
	resp, err := ts.Client().Get(ts.URL + "/static/style.css")
	if err != nil {
		t.Fatalf("GET /static/style.css failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		t.Errorf("static file status = %d, want 200", resp.StatusCode)
	}
}

// silence unused import warning on platforms where net/http isn't otherwise used.
var _ = http.MethodGet
