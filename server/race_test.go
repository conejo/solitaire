package server

import (
	"strconv"
	"sync"
	"testing"

	"github.com/solitaire/engine"
	"github.com/solitaire/games/klondike"
	"go.uber.org/zap"
)

// TestOriginalAPIRace drives the s.Games map from many goroutines using
// the ORIGINAL (pre-fix) getOrCreateGame body. Without the sync.RWMutex,
// this is guaranteed to trigger "fatal error: concurrent map read and
// map write" — both with and without the race detector.
//
// This file exists only to document the regression test for the fix
// applied to handlers.go. After the fix, the real getOrCreateGame in
// handlers.go serializes map access.
func TestOriginalAPIRace(t *testing.T) {
	logger := zap.NewNop()
	s, err := NewServer(logger)
	if err != nil {
		t.Fatal(err)
	}

	// Pre-fix body: no synchronization at all.
	getOrCreate := func(sessionID, gameType string) engine.Game {
		if gs, ok := s.Games[sessionID]; ok {
			if g, ok := gs.Game.(engine.Game); ok {
				return g
			}
		}
		var game engine.Game
		switch gameType {
		case "klondike":
			game = klondike.NewKlondikeGame(3)
		default:
			game = klondike.NewKlondikeGame(3)
		}
		game.NewGame()
		s.Games[sessionID] = GameSession{Game: game, GameType: gameType}
		return game
	}

	const goroutines = 16
	const iterations = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sid := strconv.Itoa(i) + "-" + strconv.Itoa(j%4)
				_ = getOrCreate(sid, "klondike")
			}
		}()
	}
	wg.Wait()
}
