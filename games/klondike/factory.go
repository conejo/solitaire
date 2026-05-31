package klondike

import "github.com/solitaire/engine"

// GameFactory creates new Klondike games.
type GameFactory struct{}

// New creates a new Klondike game with the default 3-card draw.
func (f *GameFactory) New() engine.Game {
	return NewKlondikeGame(3)
}

// NewWithDrawCount creates a new Klondike game with a specified draw count.
func (f *GameFactory) NewWithDrawCount(drawCount int) engine.Game {
	return NewKlondikeGame(drawCount)
}

// DefaultFactory is the default game factory instance.
var DefaultFactory = &GameFactory{}
