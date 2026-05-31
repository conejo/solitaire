package klondike

import (
	"errors"
	"fmt"

	"github.com/solitaire/engine"
)

// KlondikeGame implements the engine.Game interface for Klondike Solitaire.
type KlondikeGame struct {
	Tableau      [7]engine.Pile
	Foundation   [4]engine.Pile
	Stock        engine.Pile
	Waste        engine.Pile
	Selected     *engine.Selection
	DrawCount    int
	OneClickMove bool
	history      []engine.GameState
}

// NewKlondikeGame creates a new Klondike game with the specified draw count.
func NewKlondikeGame(drawCount int) engine.Game {
	return &KlondikeGame{
		DrawCount:    drawCount,
		OneClickMove: true,
	}
}

// NewGame initializes a new game by shuffling and dealing cards.
func (g *KlondikeGame) NewGame() error {
	// Reset all piles
	for i := range g.Tableau {
		g.Tableau[i] = engine.Pile{}
	}
	for i := range g.Foundation {
		g.Foundation[i] = engine.Pile{}
	}
	g.Stock = engine.Pile{}
	g.Waste = engine.Pile{}
	g.Selected = nil
	g.history = nil // clear history on new game

	// Create and shuffle deck
	deck := engine.NewDeck()
	deck = engine.Shuffle(deck)

	// Deal tableau: pile i gets i+1 cards, top card face-up
	cardIndex := 0
	for i := 0; i < 7; i++ {
		for j := 0; j <= i; j++ {
			if cardIndex >= len(deck) {
				return errors.New("not enough cards to deal")
			}
			card := deck[cardIndex]
			cardIndex++
			if j == i {
				card.FaceUp = true
			}
			g.Tableau[i].Push(card)
		}
	}

	// Remaining cards go to stock
	for cardIndex < len(deck) {
		g.Stock.Push(deck[cardIndex])
		cardIndex++
	}

	return nil
}

// DrawFromStock moves cards from stock to waste.
func (g *KlondikeGame) DrawFromStock() error {
	g.saveState()

	// If stock is empty, return waste to stock
	if g.Stock.IsEmpty() {
		if g.Waste.IsEmpty() {
			return errors.New("no cards to draw")
		}
		// Move all waste cards back to stock face-down
		for !g.Waste.IsEmpty() {
			card, _ := g.Waste.Pop()
			card.FaceUp = false
			g.Stock.Push(card)
		}
		return nil
	}

	// Draw up to DrawCount cards from stock to waste
	for i := 0; i < g.DrawCount; i++ {
		if g.Stock.IsEmpty() {
			break
		}
		card, _ := g.Stock.Pop()
		card.FaceUp = true
		g.Waste.Push(card)
	}

	return nil
}

// Select handles clicking on a card or pile to select it.
func (g *KlondikeGame) Select(pileType string, pileIndex int, cardIndex int) error {
	// If we already have a selection, try to move instead
	if g.Selected != nil {
		return g.MoveTo(pileType, pileIndex)
	}

	var cards []engine.Card
	var actualCardIndex int

	switch pileType {
	case "tableau":
		if pileIndex < 0 || pileIndex >= 7 {
			return errors.New("invalid tableau index")
		}
		pile := g.Tableau[pileIndex]
		if pile.IsEmpty() {
			return errors.New("empty pile")
		}
		if cardIndex < 0 || cardIndex >= pile.Size() {
			return errors.New("invalid card index")
		}
		card, _ := pile.CardAt(cardIndex)
		if !card.FaceUp {
			return errors.New("cannot select face-down card")
		}
		// Select this card and all face-up cards above it
		cards = pile.Cards[cardIndex:]
		actualCardIndex = cardIndex

	case "waste":
		if g.Waste.IsEmpty() {
			return errors.New("empty waste pile")
		}
		card, _ := g.Waste.Peek()
		cards = []engine.Card{card}
		actualCardIndex = g.Waste.Size() - 1

	case "foundation":
		if pileIndex < 0 || pileIndex >= 4 {
			return errors.New("invalid foundation index")
		}
		if g.Foundation[pileIndex].IsEmpty() {
			return errors.New("empty foundation")
		}
		card, _ := g.Foundation[pileIndex].Peek()
		cards = []engine.Card{card}
		actualCardIndex = g.Foundation[pileIndex].Size() - 1

	default:
		return fmt.Errorf("unknown pile type: %s", pileType)
	}

	// One-click move: try to auto-move the clicked card(s)
	if g.OneClickMove {
		movingCard := cards[0]
		// Try foundation first (only for single cards)
		if len(cards) == 1 {
			for i := range g.Foundation {
				if g.isValidFoundationMove(movingCard, &g.Foundation[i]) {
					g.saveState()
					g.removeFromSource(pileType, pileIndex, 1)
					g.Foundation[i].Push(movingCard)
					g.flipTopIfNeeded(pileType, pileIndex)
					return nil
				}
			}
		}
		// Try tableau (works for single cards and valid sequences)
		for i := range g.Tableau {
			// Only skip same pile if source is also tableau
			if pileType == "tableau" && i == pileIndex {
				continue
			}
			if g.isValidTableauMove(movingCard, &g.Tableau[i]) {
				g.saveState()
				g.removeFromSource(pileType, pileIndex, len(cards))
				g.Tableau[i].AddCards(cards)
				g.flipTopIfNeeded(pileType, pileIndex)
				return nil
			}
		}
	}

	g.Selected = &engine.Selection{
		PileType:  pileType,
		PileIndex: pileIndex,
		CardIndex: actualCardIndex,
		Cards:     cards,
	}

	return nil
}

// MoveTo attempts to move the selected cards to the target pile.
func (g *KlondikeGame) MoveTo(pileType string, pileIndex int) error {
	if g.Selected == nil {
		return errors.New("no selection")
	}

	if g.Selected.PileType == pileType && g.Selected.PileIndex == pileIndex {
		// Clicked same pile — deselect
		g.Selected = nil
		return nil
	}

	sourceType := g.Selected.PileType
	sourceIndex := g.Selected.PileIndex
	cards := g.Selected.Cards

	if len(cards) == 0 {
		g.Selected = nil
		return errors.New("no cards selected")
	}

	movingCard := cards[0]

	var success bool

	switch pileType {
	case "tableau":
		if pileIndex < 0 || pileIndex >= 7 {
			return errors.New("invalid tableau index")
		}
		targetPile := &g.Tableau[pileIndex]
		if g.isValidTableauMove(movingCard, targetPile) {
			g.saveState()
			// Remove cards from source
			g.removeFromSource(sourceType, sourceIndex, len(cards))
			// Add to target
			targetPile.AddCards(cards)
			// Flip next card face-up if needed
			g.flipTopIfNeeded(sourceType, sourceIndex)
			success = true
		}

	case "foundation":
		if pileIndex < 0 || pileIndex >= 4 {
			return errors.New("invalid foundation index")
		}
		// Can only move one card to foundation
		if len(cards) > 1 {
			g.Selected = nil
			return errors.New("can only move one card to foundation")
		}
		targetPile := &g.Foundation[pileIndex]
		if g.isValidFoundationMove(movingCard, targetPile) {
			g.saveState()
			g.removeFromSource(sourceType, sourceIndex, 1)
			targetPile.Push(movingCard)
			g.flipTopIfNeeded(sourceType, sourceIndex)
			success = true
		}

	default:
		g.Selected = nil
		return fmt.Errorf("unknown target pile type: %s", pileType)
	}

	g.Selected = nil

	if !success {
		return errors.New("invalid move")
	}

	return nil
}

// removeFromSource removes cards from the source pile.
func (g *KlondikeGame) removeFromSource(sourceType string, sourceIndex int, count int) {
	switch sourceType {
	case "tableau":
		pile := &g.Tableau[sourceIndex]
		pile.Cards = pile.Cards[:len(pile.Cards)-count]
	case "waste":
		for i := 0; i < count; i++ {
			g.Waste.Pop()
		}
	case "foundation":
		for i := 0; i < count; i++ {
			g.Foundation[sourceIndex].Pop()
		}
	}
}

// flipTopIfNeeded flips the top card face-up after a move.
func (g *KlondikeGame) flipTopIfNeeded(sourceType string, sourceIndex int) {
	if sourceType == "tableau" {
		pile := &g.Tableau[sourceIndex]
		if !pile.IsEmpty() && !pile.TopFaceUp() {
			pile.Cards[pile.Size()-1].FaceUp = true
		}
	}
}

// isValidTableauMove checks if a card can be placed on a tableau pile.
func (g *KlondikeGame) isValidTableauMove(card engine.Card, targetPile *engine.Pile) bool {
	if targetPile.IsEmpty() {
		// Only Kings can be placed on empty tableau piles
		return card.Rank == engine.King
	}

	topCard, _ := targetPile.Peek()
	// Must be opposite color and one rank lower
	if card.Suit.IsRed() == topCard.Suit.IsRed() {
		return false
	}
	return int(card.Rank) == int(topCard.Rank)-1
}

// isValidFoundationMove checks if a card can be placed on a foundation pile.
func (g *KlondikeGame) isValidFoundationMove(card engine.Card, targetPile *engine.Pile) bool {
	if targetPile.IsEmpty() {
		// Only Aces can start a foundation pile
		return card.Rank == engine.Ace
	}

	topCard, _ := targetPile.Peek()
	// Must be same suit and one rank higher
	return card.Suit == topCard.Suit && int(card.Rank) == int(topCard.Rank)+1
}

// AutoMoveToFoundation attempts to move all possible cards to foundation piles.
// It loops until no more moves are available.
func (g *KlondikeGame) AutoMoveToFoundation() error {
	g.saveState()

	moved := true
	for moved {
		moved = false

		// Try waste first
		if !g.Waste.IsEmpty() {
			card, _ := g.Waste.Peek()
			for i := range g.Foundation {
				if g.isValidFoundationMove(card, &g.Foundation[i]) {
					g.Waste.Pop()
					g.Foundation[i].Push(card)
					moved = true
					break
				}
			}
		}

		if moved {
			continue
		}

		// Try each tableau pile's top card
		for i := range g.Tableau {
			if g.Tableau[i].IsEmpty() {
				continue
			}
			card, _ := g.Tableau[i].Peek()
			if !card.FaceUp {
				continue
			}
			for j := range g.Foundation {
				if g.isValidFoundationMove(card, &g.Foundation[j]) {
					g.Tableau[i].Pop()
					g.Foundation[j].Push(card)
					// Flip next card if needed
					if !g.Tableau[i].IsEmpty() && !g.Tableau[i].TopFaceUp() {
						g.Tableau[i].Cards[g.Tableau[i].Size()-1].FaceUp = true
					}
					moved = true
					break
				}
			}
			if moved {
				break
			}
		}
	}

	return nil
}

// saveState pushes a copy of the current state onto the history stack.
func (g *KlondikeGame) saveState() {
	state := g.GetState()
	g.history = append(g.history, state)
	// Limit history to 100 moves
	if len(g.history) > 100 {
		g.history = g.history[len(g.history)-100:]
	}
}

// Undo restores the game to the previous state.
func (g *KlondikeGame) Undo() error {
	if len(g.history) == 0 {
		return errors.New("nothing to undo")
	}
	// Pop the last state
	idx := len(g.history) - 1
	state := g.history[idx]
	g.history = g.history[:idx]

	// Restore all piles from the saved state
	g.Tableau = state.Tableau
	g.Foundation = state.Foundation
	g.Stock = state.Stock
	g.Waste = state.Waste
	g.Selected = state.Selected
	g.DrawCount = state.DrawCount
	g.OneClickMove = state.OneClickMove

	return nil
}

// CanUndo returns true if there is a move to undo.
func (g *KlondikeGame) CanUndo() bool {
	return len(g.history) > 0
}
func (g *KlondikeGame) CheckWin() bool {
	for i := range g.Foundation {
		if g.Foundation[i].Size() != 13 {
			return false
		}
	}
	return true
}

// GetState returns the current game state for template rendering.
func (g *KlondikeGame) GetState() engine.GameState {
	return engine.GameState{
		Tableau:      g.Tableau,
		Foundation:   g.Foundation,
		Stock:        g.Stock,
		Waste:        g.Waste,
		Selected:     g.Selected,
		DrawCount:    g.DrawCount,
		GameType:     "klondike",
		IsWon:        g.CheckWin(),
		OneClickMove: g.OneClickMove,
	}
}

// SetDrawCount sets the number of cards to draw from stock.
func (g *KlondikeGame) SetDrawCount(n int) {
	g.DrawCount = n
}

// GetDrawCount returns the current draw count.
func (g *KlondikeGame) GetDrawCount() int {
	return g.DrawCount
}

// ClearSelection clears the current selection.
func (g *KlondikeGame) ClearSelection() {
	g.Selected = nil
}

// GetSelection returns the current selection.
func (g *KlondikeGame) GetSelection() *engine.Selection {
	return g.Selected
}

// ToggleOneClickMove toggles the one-click move feature.
func (g *KlondikeGame) ToggleOneClickMove() {
	g.OneClickMove = !g.OneClickMove
}
