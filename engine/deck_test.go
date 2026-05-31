package engine

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	if len(deck) != 52 {
		t.Errorf("NewDeck() returned %d cards, want 52", len(deck))
	}

	// Verify all cards are face down
	for i, card := range deck {
		if card.FaceUp {
			t.Errorf("card %d is face up, expected face down", i)
		}
	}

	// Verify we have exactly 13 of each suit
	suitCounts := make(map[Suit]int)
	rankCounts := make(map[Rank]int)

	for _, card := range deck {
		suitCounts[card.Suit]++
		rankCounts[card.Rank]++
	}

	for s := Hearts; s <= Spades; s++ {
		if suitCounts[s] != 13 {
			t.Errorf("suit %s count = %d, want 13", s.String(), suitCounts[s])
		}
	}

	for r := Ace; r <= King; r++ {
		if rankCounts[r] != 4 {
			t.Errorf("rank %s count = %d, want 4", r.String(), rankCounts[r])
		}
	}
}

func TestShuffle(t *testing.T) {
	deck := NewDeck()
	shuffled := Shuffle(deck)

	if len(shuffled) != 52 {
		t.Errorf("Shuffle() returned %d cards, want 52", len(shuffled))
	}

	// Verify original deck is unchanged
	if len(deck) != 52 {
		t.Errorf("original deck modified, length = %d", len(deck))
	}

	// Verify all cards are still present (same multiset)
	originalCounts := make(map[Card]int)
	shuffledCounts := make(map[Card]int)

	for _, c := range deck {
		originalCounts[c]++
	}
	for _, c := range shuffled {
		shuffledCounts[c]++
	}

	for card, count := range originalCounts {
		if shuffledCounts[card] != count {
			t.Errorf("card %v count mismatch after shuffle", card)
		}
	}

	// Note: There's a tiny chance the shuffle returns the same order,
	// but it's extremely unlikely with 52 cards.
	// We won't test for order change to avoid flaky tests.
}

func TestShuffleEmptyDeck(t *testing.T) {
	empty := []Card{}
	shuffled := Shuffle(empty)
	if len(shuffled) != 0 {
		t.Errorf("Shuffle(empty) returned %d cards, want 0", len(shuffled))
	}
}
