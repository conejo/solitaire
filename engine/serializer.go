package engine

// CardColor returns the Tailwind text color class for a card.
func CardColor(c Card) string {
	if c.Suit.IsRed() {
		return "text-red-600"
	}
	return "text-slate-900"
}

// CardSymbol returns the Unicode suit symbol.
func CardSymbol(c Card) string {
	return c.Suit.Symbol()
}

// CardRank returns the string representation of a card's rank.
func CardRank(c Card) string {
	return c.Rank.String()
}

// IsSelected checks if a specific card is part of the current selection.
func IsSelected(sel *Selection, pileType string, pileIndex int, cardIndex int) bool {
	if sel == nil {
		return false
	}
	return sel.PileType == pileType && sel.PileIndex == pileIndex && sel.CardIndex == cardIndex
}

// FoundationSuit returns the expected suit for a foundation pile index.
func FoundationSuit(index int) Suit {
	// In Klondike, foundations don't have a fixed suit until the first card is placed.
	// This is a placeholder for games that do assign suits to foundations.
	return Suit(index)
}

// PileID generates a unique ID string for a pile.
func PileID(pileType string, index int) string {
	return pileType + "-" + string(rune('0'+index))
}
