package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/solitaire/engine"
	"github.com/solitaire/games/klondike"
	"go.uber.org/zap"
)

// getSessionID extracts or creates a session ID from the request.
func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	// Generate a new random session ID
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// setSessionCookie sets the session ID cookie on the response.
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	})
}

// getOrCreateGame retrieves or creates a game for the session.
func (s *Server) getOrCreateGame(sessionID string, gameType string) engine.Game {
	if gs, ok := s.Games[sessionID]; ok {
		if g, ok := gs.Game.(engine.Game); ok {
			return g
		}
	}

	// Create new game
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

// handleIndex renders the main game page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	setSessionCookie(w, sessionID)

	gameType := r.URL.Query().Get("game")
	if gameType == "" {
		gameType = "klondike"
	}

	game := s.getOrCreateGame(sessionID, gameType)
	state := game.GetState()

	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		s.Logger.Error("failed to render index", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	setSessionCookie(w, sessionID)

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

	game.NewGame()
	s.Games[sessionID] = GameSession{Game: game, GameType: gameType}

	state := game.GetState()
	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Templates.ExecuteTemplate(w, "board.html", data); err != nil {
		s.Logger.Error("failed to render board", zap.Error(err), zap.String("session_id", sessionID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	game := s.getOrCreateGame(sessionID, "")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.Logger.Warn("failed to parse form", zap.Error(err), zap.String("session_id", sessionID))
		state := game.GetState()
		state.Error = "failed to parse form"
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
			renderBoard(w, s.Templates, state, "klondike")
			return
		}
	}

	state := game.GetState()
	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("select/move",
		zap.String("session_id", sessionID),
		zap.String("pile_type", pileType),
		zap.Int("pile_index", pileIndex),
		zap.Int("card_index", cardIndex),
		zap.Bool("has_selection", game.GetSelection() != nil),
		zap.Duration("duration", time.Since(start)),
	)
}

// handleStock handles drawing from the stock pile.
func (s *Server) handleStock(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sessionID := getSessionID(r)
	game := s.getOrCreateGame(sessionID, "")

	err := game.DrawFromStock()
	if err != nil {
		s.Logger.Warn("draw from stock failed", zap.Error(err), zap.String("session_id", sessionID))
		state := game.GetState()
		state.Error = err.Error()
		renderBoard(w, s.Templates, state, "klondike")
		return
	}

	state := game.GetState()
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
	game := s.getOrCreateGame(sessionID, "")

	game.AutoMoveToFoundation()

	state := game.GetState()
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
	game := s.getOrCreateGame(sessionID, "")

	currentDraw := game.GetDrawCount()
	newDraw := 3
	if currentDraw == 3 {
		newDraw = 1
	}
	game.SetDrawCount(newDraw)

	state := game.GetState()
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
	game := s.getOrCreateGame(sessionID, "")

	game.ToggleOneClickMove()

	state := game.GetState()
	renderBoard(w, s.Templates, state, "klondike")

	s.Logger.Info("toggle one-click move",
		zap.String("session_id", sessionID),
		zap.Bool("one_click_move", game.GetState().OneClickMove),
		zap.Duration("duration", time.Since(start)),
	)
}

// renderBoard renders the board template with the given state.
func renderBoard(w http.ResponseWriter, tmpl *template.Template, state engine.GameState, gameType string) {
	data := struct {
		State    engine.GameState
		GameType string
	}{
		State:    state,
		GameType: gameType,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "board.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
