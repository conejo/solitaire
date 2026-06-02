package klondike

import (
	"testing"

	"github.com/solitaire/engine"
)

func TestMoveKingFromWasteToEmptyTableau(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	// Clear tableau pile 0
	g.Tableau[0] = engine.Pile{}

	// Put a King of Spades on the waste pile
	g.Waste = engine.Pile{}
	g.Waste.Push(engine.Card{Suit: engine.Spades, Rank: engine.King, FaceUp: true})

	// Select the waste card (one-click move should auto-move it)
	err := g.Select("waste", 0, 0)
	if err != nil {
		t.Fatalf("Select(waste) failed: %v", err)
	}

	// Verify the King moved to tableau pile 0
	if g.Tableau[0].IsEmpty() {
		t.Fatal("expected King to be moved to tableau pile 0, but pile is empty")
	}

	topCard, _ := g.Tableau[0].Peek()
	if topCard.Rank != engine.King {
		t.Errorf("expected King on tableau pile 0, got %s", topCard.Rank.String())
	}
	if topCard.Suit != engine.Spades {
		t.Errorf("expected Spades, got %s", topCard.Suit.String())
	}

	// Verify waste is now empty
	if !g.Waste.IsEmpty() {
		t.Error("expected waste to be empty after move")
	}
}

func TestMoveKingFromWasteToEmptyTableauWithOneClickOff(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Clear tableau pile 0
	g.Tableau[0] = engine.Pile{}

	// Put a King of Hearts on the waste pile
	g.Waste = engine.Pile{}
	g.Waste.Push(engine.Card{Suit: engine.Hearts, Rank: engine.King, FaceUp: true})

	// Select the waste card
	err := g.Select("waste", 0, 0)
	if err != nil {
		t.Fatalf("Select(waste) failed: %v", err)
	}

	// Verify selection was made
	if g.Selected == nil {
		t.Fatal("expected card to be selected when OneClickMove is off")
	}

	// Now move to tableau pile 0
	err = g.MoveTo("tableau", 0)
	if err != nil {
		t.Fatalf("MoveTo(tableau, 0) failed: %v", err)
	}

	// Verify the King moved
	if g.Tableau[0].IsEmpty() {
		t.Fatal("expected King to be moved to tableau pile 0")
	}

	topCard, _ := g.Tableau[0].Peek()
	if topCard.Rank != engine.King {
		t.Errorf("expected King, got %s", topCard.Rank.String())
	}
}

func TestAutoMoveWasteToFoundation(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	// Put an Ace of Diamonds on the waste pile
	g.Waste = engine.Pile{}
	g.Waste.Push(engine.Card{Suit: engine.Diamonds, Rank: engine.Ace, FaceUp: true})

	// Select the waste card (one-click should auto-move to foundation)
	err := g.Select("waste", 0, 0)
	if err != nil {
		t.Fatalf("Select(waste) failed: %v", err)
	}

	// Verify the Ace moved to a foundation pile
	found := false
	for i := range g.Foundation {
		if !g.Foundation[i].IsEmpty() {
			topCard, _ := g.Foundation[i].Peek()
			if topCard.Rank == engine.Ace && topCard.Suit == engine.Diamonds {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected Ace of Diamonds to be moved to foundation")
	}
}

func TestCannotMoveNonKingToEmptyTableau(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Clear tableau pile 0
	g.Tableau[0] = engine.Pile{}

	// Put a Queen on the waste pile
	g.Waste = engine.Pile{}
	g.Waste.Push(engine.Card{Suit: engine.Hearts, Rank: engine.Queen, FaceUp: true})

	// Select the waste card
	g.Select("waste", 0, 0)

	// Try to move to empty tableau
	err := g.MoveTo("tableau", 0)
	if err == nil {
		t.Error("expected error when moving non-King to empty tableau")
	}
}

func TestNewGameDealsCorrectly(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	err := g.NewGame()
	if err != nil {
		t.Fatalf("NewGame() failed: %v", err)
	}

	// Tableau should have 1+2+3+4+5+6+7 = 28 cards
	totalTableau := 0
	for i := range g.Tableau {
		totalTableau += g.Tableau[i].Size()
		// Top card should be face-up
		if !g.Tableau[i].TopFaceUp() {
			t.Errorf("tableau pile %d top card is face-down", i)
		}
		// Pile i should have i+1 cards
		if g.Tableau[i].Size() != i+1 {
			t.Errorf("tableau pile %d has %d cards, want %d", i, g.Tableau[i].Size(), i+1)
		}
	}
	if totalTableau != 28 {
		t.Errorf("total tableau cards = %d, want 28", totalTableau)
	}

	// Stock should have 24 cards (52 - 28)
	if g.Stock.Size() != 24 {
		t.Errorf("stock has %d cards, want 24", g.Stock.Size())
	}

	// Waste should be empty
	if !g.Waste.IsEmpty() {
		t.Error("waste should be empty after NewGame")
	}

	// Foundations should be empty
	for i := range g.Foundation {
		if !g.Foundation[i].IsEmpty() {
			t.Errorf("foundation pile %d should be empty", i)
		}
	}
}

func TestDrawFromStock(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	stockBefore := g.Stock.Size()
	err := g.DrawFromStock()
	if err != nil {
		t.Fatalf("DrawFromStock() failed: %v", err)
	}

	if g.Stock.Size() != stockBefore-1 {
		t.Errorf("stock size = %d, want %d", g.Stock.Size(), stockBefore-1)
	}
	if g.Waste.Size() != 1 {
		t.Errorf("waste size = %d, want 1", g.Waste.Size())
	}

	// Drawn card should be face-up
	topCard, _ := g.Waste.Peek()
	if !topCard.FaceUp {
		t.Error("drawn card should be face-up")
	}
}

func TestDrawFromStockThreeCards(t *testing.T) {
	g := NewKlondikeGame(3).(*KlondikeGame)
	g.NewGame()

	err := g.DrawFromStock()
	if err != nil {
		t.Fatalf("DrawFromStock() failed: %v", err)
	}

	if g.Waste.Size() != 3 {
		t.Errorf("waste size = %d, want 3", g.Waste.Size())
	}
}

func TestRecycleWasteToStock(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	// Draw all cards from stock to waste
	for !g.Stock.IsEmpty() {
		g.DrawFromStock()
	}

	if !g.Stock.IsEmpty() {
		t.Fatal("stock should be empty")
	}
	if g.Waste.IsEmpty() {
		t.Fatal("waste should have cards")
	}

	wasteSize := g.Waste.Size()

	// Now recycle
	err := g.DrawFromStock()
	if err != nil {
		t.Fatalf("DrawFromStock() recycle failed: %v", err)
	}

	if g.Stock.Size() != wasteSize {
		t.Errorf("stock size = %d, want %d", g.Stock.Size(), wasteSize)
	}
	if !g.Waste.IsEmpty() {
		t.Error("waste should be empty after recycle")
	}

	// Recycled cards should be face-down
	topCard, _ := g.Stock.Peek()
	if topCard.FaceUp {
		t.Error("recycled stock cards should be face-down")
	}
}

func TestCannotDrawFromEmptyStockAndWaste(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	// Empty both stock and waste directly
	g.Stock = engine.Pile{}
	g.Waste = engine.Pile{}

	err := g.DrawFromStock()
	if err == nil {
		t.Error("expected error when drawing from empty stock and waste")
	}
}

func TestSelectFaceDownCard(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	// Try to select the bottom (face-down) card of tableau pile 6
	err := g.Select("tableau", 6, 0)
	if err == nil {
		t.Error("expected error when selecting face-down card")
	}
}

func TestSelectEmptyPile(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	// Clear tableau pile 0 and try to select
	g.Tableau[0] = engine.Pile{}
	err := g.Select("tableau", 0, 0)
	if err == nil {
		t.Error("expected error when selecting empty pile")
	}
}

func TestMoveToFoundation(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Put Ace of Spades on tableau pile 0
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.Ace, FaceUp: true})

	// Select it
	g.Select("tableau", 0, 0)

	// Move to foundation 0
	err := g.MoveTo("foundation", 0)
	if err != nil {
		t.Fatalf("MoveTo(foundation) failed: %v", err)
	}

	if g.Foundation[0].IsEmpty() {
		t.Fatal("expected Ace on foundation 0")
	}

	topCard, _ := g.Foundation[0].Peek()
	if topCard.Rank != engine.Ace || topCard.Suit != engine.Spades {
		t.Errorf("expected Ace of Spades, got %s of %s", topCard.Rank.String(), topCard.Suit.String())
	}

	// Tableau should be empty
	if !g.Tableau[0].IsEmpty() {
		t.Error("tableau pile 0 should be empty")
	}
}

func TestMoveToFoundationSameSuitSequence(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Foundation has Ace of Hearts
	g.Foundation[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Ace, FaceUp: true})

	// Tableau has 2 of Hearts
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Two, FaceUp: true})

	g.Select("tableau", 0, 0)
	err := g.MoveTo("foundation", 0)
	if err != nil {
		t.Fatalf("MoveTo(foundation) failed: %v", err)
	}

	topCard, _ := g.Foundation[0].Peek()
	if topCard.Rank != engine.Two {
		t.Errorf("expected Two, got %s", topCard.Rank.String())
	}
}

func TestCannotMoveWrongSuitToFoundation(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Foundation has Ace of Spades
	g.Foundation[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.Ace, FaceUp: true})

	// Tableau has 2 of Hearts
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Two, FaceUp: true})

	g.Select("tableau", 0, 0)
	err := g.MoveTo("foundation", 0)
	if err == nil {
		t.Error("expected error when moving wrong suit to foundation")
	}
}

func TestMoveToTableauAlternatingColor(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0 has black King
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.King, FaceUp: true})

	// Tableau 1 has red Queen
	g.Tableau[1] = engine.Pile{}
	g.Tableau[1].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Queen, FaceUp: true})

	g.Select("tableau", 1, 0)
	err := g.MoveTo("tableau", 0)
	if err != nil {
		t.Fatalf("MoveTo(tableau) failed: %v", err)
	}

	if g.Tableau[0].Size() != 2 {
		t.Errorf("tableau 0 size = %d, want 2", g.Tableau[0].Size())
	}
}

func TestCannotMoveSameColorToTableau(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0 has black King
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.King, FaceUp: true})

	// Tableau 1 has black Queen
	g.Tableau[1] = engine.Pile{}
	g.Tableau[1].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Queen, FaceUp: true})

	g.Select("tableau", 1, 0)
	err := g.MoveTo("tableau", 0)
	if err == nil {
		t.Error("expected error when moving same color to tableau")
	}
}

func TestAutoMoveToFoundationMultiple(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	// Set up: waste has Ace, tableau has 2 of same suit
	g.Waste.Push(engine.Card{Suit: engine.Diamonds, Rank: engine.Ace, FaceUp: true})
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Diamonds, Rank: engine.Two, FaceUp: true})

	g.AutoMoveToFoundation()

	// Both should have moved to foundation
	if g.Foundation[0].IsEmpty() {
		t.Fatal("expected cards on foundation")
	}

	// Foundation should have Two on top
	topCard, _ := g.Foundation[0].Peek()
	if topCard.Rank != engine.Two {
		t.Errorf("expected Two on foundation, got %s", topCard.Rank.String())
	}

	// Waste should be empty
	if !g.Waste.IsEmpty() {
		t.Error("waste should be empty")
	}

	// Tableau should be empty
	if !g.Tableau[0].IsEmpty() {
		t.Error("tableau should be empty")
	}
}

func TestCheckWinFalse(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.NewGame()

	if g.CheckWin() {
		t.Error("CheckWin() should be false for new game")
	}
}

func TestCheckWinTrue(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	// Fill all foundations with 13 cards each
	for s := engine.Hearts; s <= engine.Spades; s++ {
		for r := engine.Ace; r <= engine.King; r++ {
			g.Foundation[int(s)].Push(engine.Card{Suit: s, Rank: r, FaceUp: true})
		}
	}

	if !g.CheckWin() {
		t.Error("CheckWin() should be true when all foundations are complete")
	}
}

func TestToggleOneClickMove(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	if !g.OneClickMove {
		t.Error("OneClickMove should default to true")
	}

	g.ToggleOneClickMove()
	if g.OneClickMove {
		t.Error("OneClickMove should be false after toggle")
	}

	g.ToggleOneClickMove()
	if !g.OneClickMove {
		t.Error("OneClickMove should be true after second toggle")
	}
}

func TestSetDrawCount(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)

	if g.GetDrawCount() != 1 {
		t.Errorf("draw count = %d, want 1", g.GetDrawCount())
	}

	g.SetDrawCount(3)
	if g.GetDrawCount() != 3 {
		t.Errorf("draw count = %d, want 3", g.GetDrawCount())
	}
}

func TestDeselectByClickingSamePile(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Put a card on tableau
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Ace, FaceUp: true})

	// Select it
	g.Select("tableau", 0, 0)
	if g.Selected == nil {
		t.Fatal("expected selection")
	}

	// Click same pile again
	err := g.MoveTo("tableau", 0)
	if err != nil {
		t.Fatalf("MoveTo same pile failed: %v", err)
	}

	if g.Selected != nil {
		t.Error("expected selection to be cleared")
	}
}

func TestMoveMultipleCardsFromTableau(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0: King (face down), Queen, Jack (all face up, valid sequence)
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.King, FaceUp: false})
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Queen, FaceUp: true})
	g.Tableau[0].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Jack, FaceUp: true})

	// Tableau 1: empty
	g.Tableau[1] = engine.Pile{}

	// Select Queen (index 1) — should select Queen and Jack
	g.Select("tableau", 0, 1)
	if g.Selected == nil {
		t.Fatal("expected selection")
	}
	if len(g.Selected.Cards) != 2 {
		t.Errorf("selected %d cards, want 2", len(g.Selected.Cards))
	}

	// Move to empty tableau 1 — should fail because only King can go on empty
	err := g.MoveTo("tableau", 1)
	if err == nil {
		t.Error("expected error moving non-King sequence to empty tableau")
	}
}

func TestFlipTopCardAfterMove(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0: face-down King, face-up Ace
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Spades, Rank: engine.King, FaceUp: false})
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Ace, FaceUp: true})

	// Move Ace to foundation
	g.Select("tableau", 0, 1)
	g.MoveTo("foundation", 0)

	// King should now be face-up
	if !g.Tableau[0].TopFaceUp() {
		t.Error("expected top card to be flipped face-up after move")
	}

	topCard, _ := g.Tableau[0].Peek()
	if topCard.Rank != engine.King {
		t.Errorf("expected King on top, got %s", topCard.Rank.String())
	}
	if !topCard.FaceUp {
		t.Error("expected King to be face-up")
	}
}

func TestGetState(t *testing.T) {
	g := NewKlondikeGame(3).(*KlondikeGame)
	g.NewGame()

	state := g.GetState()

	if state.GameType != "klondike" {
		t.Errorf("game type = %q, want klondike", state.GameType)
	}
	if state.DrawCount != 3 {
		t.Errorf("draw count = %d, want 3", state.DrawCount)
	}
	if state.IsWon {
		t.Error("IsWon should be false for new game")
	}
	if !state.OneClickMove {
		t.Error("OneClickMove should be true")
	}
}

func TestClearSelection(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Ace, FaceUp: true})

	g.Select("tableau", 0, 0)
	if g.Selected == nil {
		t.Fatal("expected selection")
	}

	g.ClearSelection()
	if g.Selected != nil {
		t.Error("expected selection to be cleared")
	}
}

func TestGetSelection(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	if g.GetSelection() != nil {
		t.Error("GetSelection should return nil when nothing selected")
	}

	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Ace, FaceUp: true})
	g.Select("tableau", 0, 0)

	if g.GetSelection() == nil {
		t.Error("GetSelection should return selection after Select")
	}
}

// TestUndoDeepCopyTableau verifies that the undo history deep-copies pile
// slices. Previously saveState() relied on Go's value-copy of GameState, but
// [N]Pile contains []Card slices whose backing arrays were shared with the
// live game. Subsequent mutations (e.g. AddCards via append) could overwrite
// the snapshot, so Undo would restore corrupted state. This test exercises
// that exact path: make a move, then mutate the destination pile, then undo
// and assert the restored state matches the pre-move state.
func TestUndoDeepCopyTableau(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0: [5♣] face-up (target)
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Five, FaceUp: true})

	// Tableau 1: [4♦, 3♣] face-up (valid alt-color sequence to move).
	g.Tableau[1] = engine.Pile{}
	g.Tableau[1].Push(engine.Card{Suit: engine.Diamonds, Rank: engine.Four, FaceUp: true})
	g.Tableau[1].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Three, FaceUp: true})

	// Move 4♦-3♣ from pile 1 onto 5♣. saveState() runs first.
	g.Select("tableau", 1, 0)
	if err := g.MoveTo("tableau", 0); err != nil {
		t.Fatalf("MoveTo(tableau, 0) failed: %v", err)
	}
	if !g.Tableau[1].IsEmpty() {
		t.Fatalf("after move, pile 1 should be empty")
	}

	// Push two NEW cards onto pile 1. The slice will reuse its existing
	// backing array (capacity >= 2) and overwrite indices 0 and 1 — the
	// exact slots the snapshot still points at. With the shallow-copy bug,
	// Undo will then read these new cards instead of the originals.
	g.Tableau[1].Push(engine.Card{Suit: engine.Spades, Rank: engine.Queen, FaceUp: true})
	g.Tableau[1].Push(engine.Card{Suit: engine.Hearts, Rank: engine.Jack, FaceUp: true})

	// Undo: pile 1 should be restored to [4♦, 3♣], not [Q♠, J♥].
	if err := g.Undo(); err != nil {
		t.Fatalf("Undo() failed: %v", err)
	}
	if g.Tableau[1].Size() != 2 {
		t.Fatalf("after undo, pile 1 size = %d, want 2", g.Tableau[1].Size())
	}
	want := []engine.Card{
		{Suit: engine.Diamonds, Rank: engine.Four, FaceUp: true},
		{Suit: engine.Clubs, Rank: engine.Three, FaceUp: true},
	}
	for i, c := range want {
		got, _ := g.Tableau[1].CardAt(i)
		if got.Suit != c.Suit || got.Rank != c.Rank || got.FaceUp != c.FaceUp {
			t.Errorf("pile 1 card %d after undo = %s%s faceUp=%v, want %s%s faceUp=%v",
				i, got.Rank, got.Suit, got.FaceUp, c.Rank, c.Suit, c.FaceUp)
		}
	}
}

// TestUndoDeepCopySelection verifies that the cards in a Selection are
// deep-copied. In Select(), the Selection.Cards slice is a sub-slice of the
// source pile's backing array, so a later mutation of the pile would
// otherwise corrupt the saved Selection.
func TestUndoDeepCopySelection(t *testing.T) {
	g := NewKlondikeGame(1).(*KlondikeGame)
	g.OneClickMove = false

	// Tableau 0: [5♣] face-up
	g.Tableau[0] = engine.Pile{}
	g.Tableau[0].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Five, FaceUp: true})

	// Tableau 1: [4♦, 3♣] face-up. After Select("tableau", 1, 0), the
	// Selection.Cards will be a sub-slice of pile 1's backing array
	// starting at index 0.
	g.Tableau[1] = engine.Pile{}
	g.Tableau[1].Push(engine.Card{Suit: engine.Diamonds, Rank: engine.Four, FaceUp: true})
	g.Tableau[1].Push(engine.Card{Suit: engine.Clubs, Rank: engine.Three, FaceUp: true})

	g.Select("tableau", 1, 0)
	if g.Selected == nil {
		t.Fatal("expected selection")
	}
	wantSel := append([]engine.Card(nil), g.Selected.Cards...)

	// Move to pile 0. saveState() ran first and (with the bug) captured a
	// Selection whose Cards share the backing array with pile 1.
	if err := g.MoveTo("tableau", 0); err != nil {
		t.Fatalf("MoveTo(tableau, 0) failed: %v", err)
	}

	// Now overwrite the source pile's backing array. Push a card onto pile
	// 1 — since capacity is still >= 2, the append will reuse the existing
	// array, overwriting index 0. (Index 1 may or may not be touched, but
	// the snapshot's Selection.Cards[0] would now read the new card.)
	g.Tableau[1].Push(engine.Card{Suit: engine.Spades, Rank: engine.Queen, FaceUp: true})

	// Undo: the restored Selection should still hold the original 4♦.
	if err := g.Undo(); err != nil {
		t.Fatalf("Undo() failed: %v", err)
	}
	sel := g.GetSelection()
	if sel == nil {
		t.Fatal("expected restored selection after Undo")
	}
	if len(sel.Cards) != len(wantSel) {
		t.Fatalf("after undo, sel cards = %d, want %d", len(sel.Cards), len(wantSel))
	}
	if sel.Cards[0] != wantSel[0] {
		t.Errorf("sel.Cards[0] = %s%s, want %s%s",
			sel.Cards[0].Rank, sel.Cards[0].Suit,
			wantSel[0].Rank, wantSel[0].Suit)
	}
}
