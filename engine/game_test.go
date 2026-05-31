package engine

import "testing"

func TestGameStateDefaults(t *testing.T) {
	gs := GameState{}

	if gs.DrawCount != 0 {
		t.Errorf("GameState.DrawCount default = %d, want 0", gs.DrawCount)
	}
	if gs.IsWon {
		t.Error("GameState.IsWon default = true")
	}
	if gs.OneClickMove {
		t.Error("GameState.OneClickMove default = true")
	}
}

func TestSelectionFields(t *testing.T) {
	sel := Selection{
		PileType:  "tableau",
		PileIndex: 2,
		CardIndex: 1,
		Cards: []Card{
			{Suit: Hearts, Rank: Ace},
		},
	}

	if sel.PileType != "tableau" {
		t.Errorf("Selection.PileType = %q, want tableau", sel.PileType)
	}
	if sel.PileIndex != 2 {
		t.Errorf("Selection.PileIndex = %d, want 2", sel.PileIndex)
	}
	if sel.CardIndex != 1 {
		t.Errorf("Selection.CardIndex = %d, want 1", sel.CardIndex)
	}
	if len(sel.Cards) != 1 {
		t.Errorf("len(Selection.Cards) = %d, want 1", len(sel.Cards))
	}
}
