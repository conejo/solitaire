package engine

// Pile represents a stack of cards (tableau, foundation, stock, waste, etc.)
type Pile struct {
	Cards []Card
}

func (p *Pile) Push(c Card) {
	p.Cards = append(p.Cards, c)
}

func (p *Pile) Pop() (Card, bool) {
	if len(p.Cards) == 0 {
		return Card{}, false
	}
	c := p.Cards[len(p.Cards)-1]
	p.Cards = p.Cards[:len(p.Cards)-1]
	return c, true
}

func (p Pile) Peek() (Card, bool) {
	if len(p.Cards) == 0 {
		return Card{}, false
	}
	return p.Cards[len(p.Cards)-1], true
}

func (p Pile) IsEmpty() bool {
	return len(p.Cards) == 0
}

func (p Pile) Size() int {
	return len(p.Cards)
}

func (p Pile) TopFaceUp() bool {
	if len(p.Cards) == 0 {
		return false
	}
	return p.Cards[len(p.Cards)-1].FaceUp
}

func (p *Pile) FlipTop() {
	if len(p.Cards) > 0 {
		p.Cards[len(p.Cards)-1].FaceUp = true
	}
}

func (p Pile) AllFaceUp() bool {
	for _, c := range p.Cards {
		if !c.FaceUp {
			return false
		}
	}
	return true
}

func (p Pile) FaceUpCards() []Card {
	var result []Card
	for _, c := range p.Cards {
		if c.FaceUp {
			result = append(result, c)
		}
	}
	return result
}

func (p Pile) FaceUpCount() int {
	count := 0
	for i := len(p.Cards) - 1; i >= 0; i-- {
		if p.Cards[i].FaceUp {
			count++
		} else {
			break
		}
	}
	return count
}

func (p Pile) CardAt(index int) (Card, bool) {
	if index < 0 || index >= len(p.Cards) {
		return Card{}, false
	}
	return p.Cards[index], true
}

func (p *Pile) RemoveFrom(index int) []Card {
	if index < 0 || index >= len(p.Cards) {
		return nil
	}
	removed := make([]Card, len(p.Cards)-index)
	copy(removed, p.Cards[index:])
	p.Cards = p.Cards[:index]
	return removed
}

func (p *Pile) AddCards(cards []Card) {
	p.Cards = append(p.Cards, cards...)
}
