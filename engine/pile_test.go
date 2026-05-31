package engine

import "testing"

func TestPilePush(t *testing.T) {
	p := Pile{}
	p.Push(Card{Suit: Hearts, Rank: Ace})

	if p.Size() != 1 {
		t.Errorf("Pile.Size() = %d, want 1", p.Size())
	}
}

func TestPilePop(t *testing.T) {
	p := Pile{}
	card, ok := p.Pop()
	if ok {
		t.Error("Pop() on empty pile returned ok=true")
	}

	p.Push(Card{Suit: Hearts, Rank: Ace})
	card, ok = p.Pop()
	if !ok {
		t.Error("Pop() returned ok=false")
	}
	if card.Rank != Ace || card.Suit != Hearts {
		t.Errorf("Pop() returned wrong card: %v", card)
	}
	if p.Size() != 0 {
		t.Errorf("Pile.Size() = %d, want 0", p.Size())
	}
}

func TestPilePeek(t *testing.T) {
	p := Pile{}
	_, ok := p.Peek()
	if ok {
		t.Error("Peek() on empty pile returned ok=true")
	}

	p.Push(Card{Suit: Diamonds, Rank: Two})
	card, ok := p.Peek()
	if !ok {
		t.Error("Peek() returned ok=false")
	}
	if card.Rank != Two || card.Suit != Diamonds {
		t.Errorf("Peek() returned wrong card: %v", card)
	}
	if p.Size() != 1 {
		t.Errorf("Pile.Size() = %d, want 1 after Peek", p.Size())
	}
}

func TestPileIsEmpty(t *testing.T) {
	p := Pile{}
	if !p.IsEmpty() {
		t.Error("Empty pile IsEmpty() = false")
	}

	p.Push(Card{Suit: Clubs, Rank: Three})
	if p.IsEmpty() {
		t.Error("Non-empty pile IsEmpty() = true")
	}
}

func TestPileTopFaceUp(t *testing.T) {
	p := Pile{}
	if p.TopFaceUp() {
		t.Error("TopFaceUp() on empty pile = true")
	}

	p.Push(Card{Suit: Spades, Rank: Four, FaceUp: false})
	if p.TopFaceUp() {
		t.Error("TopFaceUp() on face-down card = true")
	}

	p.Push(Card{Suit: Hearts, Rank: Five, FaceUp: true})
	if !p.TopFaceUp() {
		t.Error("TopFaceUp() on face-up card = false")
	}
}

func TestPileFlipTop(t *testing.T) {
	p := Pile{}
	p.FlipTop() // should not panic on empty pile

	p.Push(Card{Suit: Hearts, Rank: Ace, FaceUp: false})
	p.FlipTop()
	if !p.TopFaceUp() {
		t.Error("FlipTop() did not flip the top card")
	}
}

func TestPileAllFaceUp(t *testing.T) {
	p := Pile{}
	if !p.AllFaceUp() {
		t.Error("AllFaceUp() on empty pile = false")
	}

	p.Push(Card{FaceUp: true})
	p.Push(Card{FaceUp: true})
	if !p.AllFaceUp() {
		t.Error("AllFaceUp() on all face-up pile = false")
	}

	p.Push(Card{FaceUp: false})
	if p.AllFaceUp() {
		t.Error("AllFaceUp() on mixed pile = true")
	}
}

func TestPileFaceUpCards(t *testing.T) {
	p := Pile{}
	p.Push(Card{Suit: Hearts, Rank: Ace, FaceUp: false})
	p.Push(Card{Suit: Diamonds, Rank: Two, FaceUp: true})
	p.Push(Card{Suit: Clubs, Rank: Three, FaceUp: true})

	faceUp := p.FaceUpCards()
	if len(faceUp) != 2 {
		t.Errorf("FaceUpCards() returned %d cards, want 2", len(faceUp))
	}
}

func TestPileFaceUpCount(t *testing.T) {
	p := Pile{}
	if p.FaceUpCount() != 0 {
		t.Errorf("FaceUpCount() on empty pile = %d, want 0", p.FaceUpCount())
	}

	p.Push(Card{FaceUp: false})
	p.Push(Card{FaceUp: false})
	p.Push(Card{FaceUp: true})
	p.Push(Card{FaceUp: true})
	if p.FaceUpCount() != 2 {
		t.Errorf("FaceUpCount() = %d, want 2", p.FaceUpCount())
	}
}

func TestPileCardAt(t *testing.T) {
	p := Pile{}
	p.Push(Card{Suit: Hearts, Rank: Ace})

	_, ok := p.CardAt(-1)
	if ok {
		t.Error("CardAt(-1) returned ok=true")
	}

	_, ok = p.CardAt(1)
	if ok {
		t.Error("CardAt(1) on single-card pile returned ok=true")
	}

	card, ok := p.CardAt(0)
	if !ok || card.Rank != Ace {
		t.Errorf("CardAt(0) returned wrong card or ok=false")
	}
}

func TestPileRemoveFrom(t *testing.T) {
	p := Pile{}
	p.Push(Card{Suit: Hearts, Rank: Ace})
	p.Push(Card{Suit: Diamonds, Rank: Two})
	p.Push(Card{Suit: Clubs, Rank: Three})

	removed := p.RemoveFrom(1)
	if len(removed) != 2 {
		t.Errorf("RemoveFrom(1) returned %d cards, want 2", len(removed))
	}
	if p.Size() != 1 {
		t.Errorf("Pile.Size() after RemoveFrom = %d, want 1", p.Size())
	}

	removed = p.RemoveFrom(-1)
	if removed != nil {
		t.Error("RemoveFrom(-1) returned non-nil")
	}

	removed = p.RemoveFrom(10)
	if removed != nil {
		t.Error("RemoveFrom(10) on small pile returned non-nil")
	}
}

func TestPileAddCards(t *testing.T) {
	p := Pile{}
	p.AddCards([]Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: Two},
	})
	if p.Size() != 2 {
		t.Errorf("Pile.Size() after AddCards = %d, want 2", p.Size())
	}
}
