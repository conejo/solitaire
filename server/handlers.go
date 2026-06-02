package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/solitaire/engine"
	"github.com/solitaire/games/klondike"
	"go.uber.org/zap"
)

// sessionCookieName is the name of the session cookie. Centralized so it's
// easy to grep and to change in one place.
const sessionCookieName = "session_id"

// sessionCookieMaxAge is how long a session cookie lives. 24 hours is a
// pragmatic default; longer-lived cookies would need a server-side
// eviction policy to keep the in-memory game map bounded.
const sessionCookieMaxAge = 24 * time.Hour

// getSessionID extracts or creates a session ID from the request.
func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	// Generate a new random session ID.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is exceptional; surface it via a timestamp
		// fallback so the request still works but it's predictable.
		// Callers should log this; the function can't log itself without
		// access to a logger.
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// setSessionCookie sets the session ID cookie on the response. The cookie
// is HttpOnly (not readable from JS — mitigates XSS theft) and SameSite=Lax
// (mitigates CSRF on the state-changing POST routes). Secure is enabled
// when the request scheme is https, or can be forced on via CookieOptions
// for proxied deployments.
func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string) {
	secure := s.CookieSecure
	if !secure && isHTTPSRequest(r) {
		// Trust the X-Forwarded-Proto header from a properly configured
		// reverse proxy. If you're not behind one, this stays false.
		secure = true
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionCookieMaxAge.Seconds()),
	})
}

// isHTTPSRequest reports whether the request came in over TLS — either
// directly (r.TLS != nil) or via a trusted reverse proxy (X-Forwarded-Proto).
func isHTTPSRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); strings.EqualFold(proto, "https") {
		return true
	}
	return false
}

// getOrCreateGame retrieves or creates a game for the session.
//
// Two locks are involved:
//  1. s.mu (read or write) protects the s.Games map itself.
//  2. The returned *GameSession's mu serializes mutations of the underlying
//     engine.Game, which is not safe for concurrent use. Callers must hold
//     that lock for the duration of any game-mutating call they make.
//
// The map stores *GameSession (not GameSession) so the per-session mutex is
// shared by all holders of the pointer — copying a sync.Mutex would be
// undefined behavior.
func (s *Server) getOrCreateGame(sessionID string, gameType string) (*GameSession, error) {
	s.mu.RLock()
	gs, ok := s.Games[sessionID]
	s.mu.RUnlock()
	if ok {
		return gs, nil
	}

	// Create new game outside the lock to keep the critical section short.
	var game engine.Game
	switch gameType {
	case "klondike":
		game = klondike.NewKlondikeGame(3)
	default:
		game = klondike.NewKlondikeGame(3)
	}

	if err := game.NewGame(); err != nil {
		return nil, err
	}

	newSession := &GameSession{Game: game, GameType: gameType}
	s.mu.Lock()
	// Re-check in case another goroutine created the same session while we
	// were dealing.
	if existing, ok := s.Games[sessionID]; ok {
		s.mu.Unlock()
		return existing, nil
	}
	s.Games[sessionID] = newSession
	s.mu.Unlock()
	return newSession, nil
}

// handleIndex renders the main game page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	s.setSessionCookie(w, r, sessionID)

	gameType := r.URL.Query().Get("game")
	if gameType == "" {
		gameType = "klondike"
	}

	gs, err := s.getOrCreateGame(sessionID, gameType)
	if err != nil {
		s.Logger.Error("failed to create game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gs.mu.Lock()
	state := gs.Game.GetState()
	gs.mu.Unlock()

	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	// Render to a buffer first so a template error doesn't trigger
	// http.Error's WriteHeader after we've already implicitly sent 200
	// (which produces a "superfluous response.WriteHeader call" warning).
	var buf bytes.Buffer
	if err := s.Templates.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		s.Logger.Error("failed to render index", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)

	s.Logger.Info("request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("session_id", sessionID),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleNewGame starts a new game.
func (s *Server) handleNewGame(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	s.setSessionCookie(w, r, sessionID)

	gameType := r.PostFormValue("game_type")
	if gameType == "" {
		gameType = "klondike"
	}

	// Create a fresh game
	var game engine.Game
	switch gameType {
	case "klondike":
		game = klondike.NewKlondikeGame(3)
	default:
		game = klondike.NewKlondikeGame(3)
	}

	if err := game.NewGame(); err != nil {
		s.Logger.Error("failed to start new game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gs := &GameSession{Game: game, GameType: gameType}
	s.mu.Lock()
	s.Games[sessionID] = gs
	s.mu.Unlock()

	gs.mu.Lock()
	state := gs.Game.GetState()
	gs.mu.Unlock()

	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	// Buffer first, then commit (avoids "superfluous WriteHeader" if
	// the template fails after we've already set headers).
	var buf bytes.Buffer
	if err := s.Templates.ExecuteTemplate(&buf, "board.html", data); err != nil {
		s.Logger.Error("failed to render board", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)

	s.Logger.Info("new game",
		zap.String("session_id", sessionID),
		zap.String("game_type", gameType),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleSelect handles card selection and movement.
func (s *Server) handleSelect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	gs, err := s.getOrCreateGame(sessionID, "")
	if err != nil {
		s.Logger.Error("failed to get game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := gs.Game

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.Logger.Warn("failed to parse form", zap.Error(err), zap.String("session_id", sessionID))
		gs.mu.Lock()
		state := game.GetState()
		state.Error = "failed to parse form"
		gs.mu.Unlock()
		renderBoard(w, s.Templates, state, "klondike")
		return
	}

	pileType := r.PostFormValue("pile_type")
	pileIndexStr := r.PostFormValue("pile_index")
	cardIndexStr := r.PostFormValue("card_index")

	pileIndex := 0
	cardIndex := 0
	fmt.Sscanf(pileIndexStr, "%d", &pileIndex)
	fmt.Sscanf(cardIndexStr, "%d", &cardIndex)

	gs.mu.Lock()
	// If no selection exists, try to select
	if game.GetSelection() == nil {
		err := game.Select(pileType, pileIndex, cardIndex)
		if err != nil {
			s.Logger.Warn("select failed",
				zap.Error(err),
				zap.String("session_id", sessionID),
				zap.String("pile_type", pileType),
				zap.Int("pile_index", pileIndex),
				zap.Int("card_index", cardIndex),
			)
			state := game.GetState()
			state.Error = err.Error()
			gs.mu.Unlock()
			renderBoard(w, s.Templates, state, "klondike")
			return
		}
	} else {
		// Try to move to target
		err := game.MoveTo(pileType, pileIndex)
		if err != nil {
			s.Logger.Warn("move failed",
				zap.Error(err),
				zap.String("session_id", sessionID),
				zap.String("target_pile", pileType),
				zap.Int("target_index", pileIndex),
			)
			state := game.GetState()
			state.Error = err.Error()
			gs.mu.Unlock()
			renderBoard(w, s.Templates, state, "klondike")
			return
		}
	}

	state := game.GetState()
	hasSelection := game.GetSelection() != nil
	gs.mu.Unlock()

	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("select/move",
		zap.String("session_id", sessionID),
		zap.String("pile_type", pileType),
		zap.Int("pile_index", pileIndex),
		zap.Int("card_index", cardIndex),
		zap.Bool("has_selection", hasSelection),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleStock handles drawing from the stock pile.
func (s *Server) handleStock(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	gs, err := s.getOrCreateGame(sessionID, "")
	if err != nil {
		s.Logger.Error("failed to get game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := gs.Game

	gs.mu.Lock()
	err = game.DrawFromStock()
	if err != nil {
		s.Logger.Warn("draw from stock failed", zap.Error(err), zap.String("session_id", sessionID))
		state := game.GetState()
		state.Error = err.Error()
		gs.mu.Unlock()
		renderBoard(w, s.Templates, state, "klondike")
		return
	}
	state := game.GetState()
	gs.mu.Unlock()

	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("draw from stock",
		zap.String("session_id", sessionID),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleFoundationAuto handles auto-moving cards to foundation.
func (s *Server) handleFoundationAuto(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	gs, err := s.getOrCreateGame(sessionID, "")
	if err != nil {
		s.Logger.Error("failed to get game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := gs.Game

	gs.mu.Lock()
	game.AutoMoveToFoundation()
	state := game.GetState()
	gs.mu.Unlock()

	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("auto move to foundation",
		zap.String("session_id", sessionID),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleDrawMode toggles between 1-card and 3-card draw.
func (s *Server) handleDrawMode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	gs, err := s.getOrCreateGame(sessionID, "")
	if err != nil {
		s.Logger.Error("failed to get game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := gs.Game

	gs.mu.Lock()
	currentDraw := game.GetDrawCount()
	newDraw := 3
	if currentDraw == 3 {
		newDraw = 1
	}
	game.SetDrawCount(newDraw)
	state := game.GetState()
	gs.mu.Unlock()

	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("toggle draw mode",
		zap.String("session_id", sessionID),
		zap.Int("old_draw", currentDraw),
		zap.Int("new_draw", newDraw),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleToggleOneClick toggles the one-click move feature.
func (s *Server) handleToggleOneClick(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	gs, err := s.getOrCreateGame(sessionID, "")
	if err != nil {
		s.Logger.Error("failed to get game", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := gs.Game

	gs.mu.Lock()
	game.ToggleOneClickMove()
	state := game.GetState()
	oneClick := state.OneClickMove
	gs.mu.Unlock()

	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("toggle one-click move",
		zap.String("session_id", sessionID),
		zap.Bool("one_click_move", oneClick),
		zap.Duration("duration", time.Since(start)),
	)
}

// renderBoard renders the board template with the given state. Renders to
// a buffer first so a template error doesn't fire http.Error's WriteHeader
// after we've already committed the 200 OK (superfluous WriteHeader).
func renderBoard(w http.ResponseWriter, tmpl *template.Template, state engine.GameState, gameType string) {
	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "board.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}
