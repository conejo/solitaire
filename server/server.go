package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"sync"

	"github.com/solitaire/engine"
	"go.uber.org/zap"
)

// Server holds the HTTP server configuration and game state.
//
// Games is guarded by mu: a read-lock is enough to look up an existing
// session, but writes (insert, replace) require the write lock. The map
// stores *GameSession pointers (not values) so that GameSession.mu is
// shared by all holders of the pointer — copying a sync.Mutex is a vet
// error and would silently corrupt the lock state.
type Server struct {
	Templates *template.Template
	Logger    *zap.Logger

	mu    sync.RWMutex
	Games map[string]*GameSession
}

// GameSession holds a game instance and its type. The mutex serializes
// mutations of Game — without it, two concurrent HTMX requests on the same
// session could interleave calls like MoveTo + DrawFromStock and corrupt
// internal state. Always access via *GameSession so the mutex is shared.
type GameSession struct {
	Game     engine.Game
	GameType string
	mu       sync.Mutex
}

// NewServer creates a new server with parsed templates.
func NewServer(logger *zap.Logger) (*Server, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	return &Server{
		Templates: tmpl,
		Games:     make(map[string]*GameSession),
		Logger:    logger,
	}, nil
}

// parseTemplates parses all HTML templates from the embedded templates directory.
func parseTemplates() (*template.Template, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"until": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call: must have even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	tmplFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		return nil, err
	}

	// Parse layout and all game templates
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(tmplFS, "*.html")
	if err != nil {
		return nil, err
	}

	// Parse klondike templates
	tmpl, err = tmpl.ParseFS(tmplFS, "klondike/*.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// Start starts the HTTP server on the given address.
func (s *Server) Start(addr string) error {
	s.Logger.Info("starting server", zap.String("addr", addr))
	return http.ListenAndServe(addr, s.Handler())
}

// Handler returns the HTTP handler for the server. Exposed so tests can
// drive routes via httptest without binding a real socket.
func (s *Server) Handler() http.Handler {
	return s.buildMux()
}

func (s *Server) buildMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Static files
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		// The embedded FS is always present at compile time; if it's
		// missing something is very wrong, so panic rather than swallow.
		panic("server: missing embedded static FS: " + err.Error())
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("POST /new", s.handleNewGame)
	mux.HandleFunc("POST /select", s.handleSelect)
	mux.HandleFunc("POST /stock", s.handleStock)
	mux.HandleFunc("POST /foundation-auto", s.handleFoundationAuto)
	mux.HandleFunc("POST /draw-mode", s.handleDrawMode)
	mux.HandleFunc("POST /toggle-one-click", s.handleToggleOneClick)

	return mux
}
