package engine

// GameState holds the serialized state of a game for template rendering.
type GameState struct {
	Tableau      [7]Pile
	Foundation   [4]Pile
	Stock        Pile
	Waste        Pile
	Selected     *Selection
	DrawCount    int
	GameType     string
	IsWon        bool
	Error        string
	OneClickMove bool
}

// Selection represents a currently selected card or pile of cards.
type Selection struct {
	PileType  string // "tableau", "foundation", "waste"
	PileIndex int    // index within the pile type
	CardIndex int    // index of the first selected card in the pile
	Cards     []Card // the selected cards
}

// Game is the interface that all solitaire variants must implement.
type Game interface {
	NewGame() error
	DrawFromStock() error
	Select(pileType string, pileIndex int, cardIndex int) error
	MoveTo(pileType string, pileIndex int) error
	AutoMoveToFoundation() error
	CheckWin() bool
	GetState() GameState
	SetDrawCount(n int)
	GetDrawCount() int
	ClearSelection()
	GetSelection() *Selection
	ToggleOneClickMove()
	Undo() error
	CanUndo() bool
}
