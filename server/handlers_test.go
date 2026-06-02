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

// TestSessionCookieHasSecurityFlags verifies that the session cookie set
// on the first response carries HttpOnly, SameSite=Lax, a non-zero MaxAge,
// and Path=/. Secure is opt-in (off by default) so a developer can use
// the cookie over plain HTTP locally.
func TestSessionCookieHasSecurityFlags(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	var found *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("session_id cookie not set")
	}
	if !found.HttpOnly {
		t.Error("cookie HttpOnly = false, want true")
	}
	if found.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want Lax", found.SameSite)
	}
	if found.Path != "/" {
		t.Errorf("cookie Path = %q, want /", found.Path)
	}
	if found.MaxAge <= 0 {
		t.Errorf("cookie MaxAge = %d, want > 0", found.MaxAge)
	}
	if found.Secure {
		t.Error("cookie Secure = true over plain HTTP, want false (default)")
	}
	if found.Value == "" {
		t.Error("cookie value is empty")
	}
}

// TestSessionCookieSecureWhenForced verifies the CookieSecure knob on
// Server takes effect: when set, every response cookie has Secure=true.
func TestSessionCookieSecureWhenForced(t *testing.T) {
	s := newTestServer(t)
	s.CookieSecure = true
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	var found *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("session_id cookie not set")
	}
	if !found.Secure {
		t.Error("CookieSecure=true but cookie.Secure = false")
	}
}

// TestSessionCookieSecureBehindProxy verifies that a request with
// X-Forwarded-Proto: https gets a Secure cookie even when the knob is
// off. This lets deployments behind a TLS-terminating reverse proxy
// (nginx, Caddy, ELB) "just work" without a config flag.
func TestSessionCookieSecureBehindProxy(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			if !c.Secure {
				t.Error("X-Forwarded-Proto=https but cookie.Secure = false")
			}
			return
		}
	}
	t.Fatal("session_id cookie not set")
}

// TestSessionCookieNotLeakableByJS is a structural assertion: HttpOnly
// is set. This is the property the rest of the test suite relies on;
// it deserves a test of its own so a future "make cookies readable from
// JS" change fails loud.
func TestSessionCookieNotLeakableByJS(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" && !c.HttpOnly {
			t.Fatalf("session_id cookie is readable from JavaScript; " +
				"this allows XSS to steal sessions. Set HttpOnly=true.")
		}
	}
}

// TestSessionIDSticky verifies the same session ID is reused on a
// subsequent request, so the cookie is doing its job of stabilizing the
// session.
func TestSessionIDSticky(t *testing.T) {
	s := newTestServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// First request — cookie should be set.
	resp1, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	var first string
	for _, c := range resp1.Cookies() {
		if c.Name == "session_id" {
			first = c.Value
		}
	}
	if first == "" {
		t.Fatal("first response had no session_id cookie")
	}

	// Second request — the client should echo the cookie back, and the
	// server should NOT issue a new one.
	req, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: first})
	resp2, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	for _, c := range resp2.Cookies() {
		if c.Name == "session_id" && c.Value != first {
			t.Errorf("session_id changed between requests: %q -> %q", first, c.Value)
		}
	}
}

// TestHttpServerHasTimeouts verifies that the *http.Server built by
// httpServer has non-zero timeouts on every field. Without these, slow
// clients can exhaust file descriptors (the "Slowloris" attack).
func TestHttpServerHasTimeouts(t *testing.T) {
	s := newTestServer(t)
	srv := s.httpServer(":0")
	if srv.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout not set")
	}
	if srv.ReadTimeout <= 0 {
		t.Error("ReadTimeout not set")
	}
	if srv.WriteTimeout <= 0 {
		t.Error("WriteTimeout not set")
	}
	if srv.IdleTimeout <= 0 {
		t.Error("IdleTimeout not set")
	}
}
