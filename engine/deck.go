package engine

import (
	"math/rand"
	"time"
)

func NewDeck() []Card {
	deck := make([]Card, 0, 52)
	for s := Hearts; s <= Spades; s++ {
		for r := Ace; r <= King; r++ {
			deck = append(deck, Card{Suit: s, Rank: r, FaceUp: false})
		}
	}
	return deck
}

func Shuffle(deck []Card) []Card {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]Card, len(deck))
	copy(shuffled, deck)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}
