package klondike

import (
	"github.com/solitaire/engine"
)

// CanSelectTableau checks if a card in a tableau pile can be selected.
func CanSelectTableau(pile engine.Pile, cardIndex int) bool {
	if pile.IsEmpty() || cardIndex < 0 || cardIndex >= pile.Size() {
		return false
	}
	card, _ := pile.CardAt(cardIndex)
	if !card.FaceUp {
		return false
	}
	// All cards from cardIndex to top must be face-up and in valid sequence
	for i := cardIndex; i < pile.Size(); i++ {
		c, _ := pile.CardAt(i)
		if !c.FaceUp {
			return false
		}
		if i > cardIndex {
			prev, _ := pile.CardAt(i - 1)
			// Must be alternating colors and descending rank
			if c.Suit.IsRed() == prev.Suit.IsRed() {
				return false
			}
			if int(c.Rank) != int(prev.Rank)-1 {
				return false
			}
		}
	}
	return true
}

// IsValidTableauSequence checks if cards form a valid descending alternating-color sequence.
func IsValidTableauSequence(cards []engine.Card) bool {
	if len(cards) == 0 {
		return false
	}
	for i := 1; i < len(cards); i++ {
		if cards[i].Suit.IsRed() == cards[i-1].Suit.IsRed() {
			return false
		}
		if int(cards[i].Rank) != int(cards[i-1].Rank)-1 {
			return false
		}
	}
	return true
}

// CanMoveToTableau checks if a card can be placed on a specific tableau pile.
func CanMoveToTableau(card engine.Card, pile engine.Pile) bool {
	if pile.IsEmpty() {
		return card.Rank == engine.King
	}
	topCard, _ := pile.Peek()
	return card.Suit.IsRed() != topCard.Suit.IsRed() && int(card.Rank) == int(topCard.Rank)-1
}

// CanMoveToFoundation checks if a card can be placed on a specific foundation pile.
func CanMoveToFoundation(card engine.Card, pile engine.Pile) bool {
	if pile.IsEmpty() {
		return card.Rank == engine.Ace
	}
	topCard, _ := pile.Peek()
	return card.Suit == topCard.Suit && int(card.Rank) == int(topCard.Rank)+1
}
