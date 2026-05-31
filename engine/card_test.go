package engine

import "testing"

func TestSuitString(t *testing.T) {
	tests := []struct {
		suit     Suit
		expected string
	}{
		{Hearts, "Hearts"},
		{Diamonds, "Diamonds"},
		{Clubs, "Clubs"},
		{Spades, "Spades"},
		{Suit(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.suit.String(); got != tt.expected {
			t.Errorf("Suit(%d).String() = %q, want %q", tt.suit, got, tt.expected)
		}
	}
}

func TestSuitSymbol(t *testing.T) {
	tests := []struct {
		suit     Suit
		expected string
	}{
		{Hearts, "♥"},
		{Diamonds, "♦"},
		{Clubs, "♣"},
		{Spades, "♠"},
		{Suit(99), "?"},
	}

	for _, tt := range tests {
		if got := tt.suit.Symbol(); got != tt.expected {
			t.Errorf("Suit(%d).Symbol() = %q, want %q", tt.suit, got, tt.expected)
		}
	}
}

func TestSuitColor(t *testing.T) {
	tests := []struct {
		suit     Suit
		expected string
	}{
		{Hearts, "red"},
		{Diamonds, "red"},
		{Clubs, "black"},
		{Spades, "black"},
	}

	for _, tt := range tests {
		if got := tt.suit.Color(); got != tt.expected {
			t.Errorf("Suit(%d).Color() = %q, want %q", tt.suit, got, tt.expected)
		}
	}
}

func TestSuitIsRed(t *testing.T) {
	tests := []struct {
		suit     Suit
		expected bool
	}{
		{Hearts, true},
		{Diamonds, true},
		{Clubs, false},
		{Spades, false},
	}

	for _, tt := range tests {
		if got := tt.suit.IsRed(); got != tt.expected {
			t.Errorf("Suit(%d).IsRed() = %v, want %v", tt.suit, got, tt.expected)
		}
	}
}

func TestRankString(t *testing.T) {
	tests := []struct {
		rank     Rank
		expected string
	}{
		{Ace, "A"},
		{Two, "2"},
		{Three, "3"},
		{Four, "4"},
		{Five, "5"},
		{Six, "6"},
		{Seven, "7"},
		{Eight, "8"},
		{Nine, "9"},
		{Ten, "10"},
		{Jack, "J"},
		{Queen, "Q"},
		{King, "K"},
		{Rank(14), "14"},
	}

	for _, tt := range tests {
		if got := tt.rank.String(); got != tt.expected {
			t.Errorf("Rank(%d).String() = %q, want %q", tt.rank, got, tt.expected)
		}
	}
}

func TestCardString(t *testing.T) {
	card := Card{Suit: Hearts, Rank: Ace}
	if got := card.String(); got != "A♥" {
		t.Errorf("Card.String() = %q, want %q", got, "A♥")
	}

	card2 := Card{Suit: Spades, Rank: King}
	if got := card2.String(); got != "K♠" {
		t.Errorf("Card.String() = %q, want %q", got, "K♠")
	}
}
