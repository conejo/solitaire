package engine

import "testing"

func TestCardColor(t *testing.T) {
	tests := []struct {
		card     Card
		expected string
	}{
		{Card{Suit: Hearts, Rank: Ace}, "text-red-600"},
		{Card{Suit: Diamonds, Rank: Two}, "text-red-600"},
		{Card{Suit: Clubs, Rank: Three}, "text-slate-900"},
		{Card{Suit: Spades, Rank: Four}, "text-slate-900"},
	}

	for _, tt := range tests {
		if got := CardColor(tt.card); got != tt.expected {
			t.Errorf("CardColor(%v) = %q, want %q", tt.card, got, tt.expected)
		}
	}
}

func TestCardSymbol(t *testing.T) {
	card := Card{Suit: Hearts, Rank: Ace}
	if got := CardSymbol(card); got != "♥" {
		t.Errorf("CardSymbol() = %q, want %q", got, "♥")
	}
}

func TestCardRank(t *testing.T) {
	card := Card{Suit: Spades, Rank: King}
	if got := CardRank(card); got != "K" {
		t.Errorf("CardRank() = %q, want %q", got, "K")
	}
}

func TestIsSelected(t *testing.T) {
	if IsSelected(nil, "tableau", 0, 0) {
		t.Error("IsSelected(nil, ...) = true")
	}

	sel := &Selection{PileType: "tableau", PileIndex: 1, CardIndex: 2}
	if !IsSelected(sel, "tableau", 1, 2) {
		t.Error("IsSelected() matching selection = false")
	}
	if IsSelected(sel, "foundation", 1, 2) {
		t.Error("IsSelected() non-matching pile type = true")
	}
	if IsSelected(sel, "tableau", 0, 2) {
		t.Error("IsSelected() non-matching pile index = true")
	}
	if IsSelected(sel, "tableau", 1, 3) {
		t.Error("IsSelected() non-matching card index = true")
	}
}

func TestFoundationSuit(t *testing.T) {
	if got := FoundationSuit(0); got != Hearts {
		t.Errorf("FoundationSuit(0) = %v, want Hearts", got)
	}
	if got := FoundationSuit(1); got != Diamonds {
		t.Errorf("FoundationSuit(1) = %v, want Diamonds", got)
	}
	if got := FoundationSuit(2); got != Clubs {
		t.Errorf("FoundationSuit(2) = %v, want Clubs", got)
	}
	if got := FoundationSuit(3); got != Spades {
		t.Errorf("FoundationSuit(3) = %v, want Spades", got)
	}
}

func TestPileID(t *testing.T) {
	if got := PileID("tableau", 0); got != "tableau-0" {
		t.Errorf("PileID() = %q, want %q", got, "tableau-0")
	}
	if got := PileID("foundation", 3); got != "foundation-3" {
		t.Errorf("PileID() = %q, want %q", got, "foundation-3")
	}
}
