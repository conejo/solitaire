package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/solitaire/engine"
	"github.com/solitaire/games/klondike"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#166534")).
			Padding(0, 1)

	cardStyle = lipgloss.NewStyle().
			Width(10).
			Height(4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFFFFF")).
			Align(lipgloss.Center, lipgloss.Top).
			Padding(0, 0).
			Bold(true)

	redCardStyle   = cardStyle.BorderForeground(lipgloss.Color("#DC2626")).Foreground(lipgloss.Color("#DC2626"))
	blackCardStyle = cardStyle.BorderForeground(lipgloss.Color("#94A3B8")).Foreground(lipgloss.Color("#0F172A")).Background(lipgloss.Color("#E2E8F0"))
	emptySlotStyle = cardStyle.BorderStyle(lipgloss.HiddenBorder())
	selectedStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#FBBF24"))
)

type model struct {
	game   engine.Game
	state  engine.GameState
	cursor int
	err    string
	help   bool
	width  int
	height int
}

func initialModel() model {
	game := klondike.NewKlondikeGame(3)
	game.NewGame()
	return model{
		game:  game,
		state: game.GetState(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			m.game.NewGame()
		case "d":
			m.game.DrawFromStock()
		case "a":
			m.game.AutoMoveToFoundation()
		case "u":
			if m.game.CanUndo() {
				m.game.Undo()
			}
		case "?":
			m.help = !m.help
		case "1", "2", "3", "4", "5", "6", "7":
			idx := int(msg.String()[0] - '1')
			if m.game.GetSelection() == nil {
				pile := m.state.Tableau[idx]
				if !pile.IsEmpty() {
					cardIdx := pile.Size() - pile.FaceUpCount()
					if cardIdx < pile.Size() {
						m.game.Select("tableau", idx, cardIdx)
					}
				}
			} else {
				m.game.MoveTo("tableau", idx)
			}
		case "w":
			if m.game.GetSelection() == nil {
				m.game.Select("waste", 0, m.state.Waste.Size()-1)
			}
		case "f":
			if m.game.GetSelection() != nil {
				for i := 0; i < 4; i++ {
					if m.game.MoveTo("foundation", i) == nil {
						break
					}
				}
			}
		case " ":
			m.game.ClearSelection()
		}
		m.state = m.game.GetState()
	}
	return m, nil
}

func (m model) View() string {
	if m.state.IsWon {
		return titleStyle.Render(" 🎉 You Win! Press 'n' for new game, 'q' to quit ") + "\n"
	}

	if m.help {
		return m.renderHelp()
	}

	var s string
	s += titleStyle.Render(" Solitaire ") + "\n\n"

	// Stock and Waste
	var stockCard, wasteCard string
	if m.state.Stock.IsEmpty() {
		stockCard = emptySlotStyle.Render("    ")
	} else {
		stockCard = cardStyle.Background(lipgloss.Color("#1E40AF")).Render("    ")
	}
	if m.state.Waste.IsEmpty() {
		wasteCard = emptySlotStyle.Render("    ")
	} else {
		top, _ := m.state.Waste.Peek()
		wasteCard = formatCard(top)
	}
	s += "Stock / Waste\n" + lipgloss.JoinHorizontal(lipgloss.Top, stockCard, wasteCard) + "\n\n"

	// Foundations
	foundationCards := make([]string, 4)
	for i := 0; i < 4; i++ {
		if m.state.Foundation[i].IsEmpty() {
			foundationCards[i] = emptySlotStyle.Render("    ")
		} else {
			top, _ := m.state.Foundation[i].Peek()
			foundationCards[i] = formatCard(top)
		}
	}
	s += "Foundations\n" + lipgloss.JoinHorizontal(lipgloss.Top, foundationCards...) + "\n\n"

	// Tableau — render 7 columns side by side
	// Calculate available height for tableau (leave room for top sections + controls)
	fixedHeight := 12 // title(2) + stock/waste(5) + foundations(5) + controls(2) + padding
	maxTableauHeight := m.height - fixedHeight
	if maxTableauHeight < 6 {
		maxTableauHeight = 6 // minimum visible tableau height
	}

	columns := make([]string, 7)
	for i := 0; i < 7; i++ {
		columns[i] = renderTableauColumn(i, m.state.Tableau[i], m.state.Selected, maxTableauHeight)
	}
	s += "Tableau\n" + lipgloss.JoinHorizontal(lipgloss.Top, columns...) + "\n"

	if m.state.Error != "" {
		s += "\nError: " + m.state.Error + "\n"
	}

	if m.state.Selected != nil {
		s += fmt.Sprintf("\nSelected: %s pile %d\n", m.state.Selected.PileType, m.state.Selected.PileIndex)
	}

	// Controls at the bottom so they're always visible
	s += "\nn=new d=draw a=auto u=undo 1-7=tab w=waste f=fnd space=clr ?=help q=quit\n"

	return s
}

func (m model) renderHelp() string {
	var s string
	s += titleStyle.Render(" Solitaire Help ") + "\n\n"
	s += "Keyboard Controls:\n\n"
	s += "  n        New game\n"
	s += "  d        Draw from stock\n"
	s += "  a        Auto-move to foundation\n"
	s += "  u        Undo last move\n"
	s += "  1-7      Select / move to tableau pile\n"
	s += "  w        Select waste card\n"
	s += "  f        Move selected card to foundation\n"
	s += "  space    Clear selection\n"
	s += "  ?        Toggle this help\n"
	s += "  q        Quit\n\n"
	s += "Press ? to return to the game.\n"
	return s
}

func renderTableauColumn(index int, pile engine.Pile, sel *engine.Selection, maxHeight int) string {
	header := lipgloss.NewStyle().
		Width(10).
		Align(lipgloss.Center).
		Bold(true).
		Render(fmt.Sprintf("%d", index+1))

	if pile.IsEmpty() {
		empty := emptySlotStyle.Render("    ")
		return lipgloss.JoinVertical(lipgloss.Left, header, empty)
	}

	var cards []string
	for j, card := range pile.Cards {
		isSelected := sel != nil && sel.PileType == "tableau" && sel.PileIndex == index && sel.CardIndex == j
		if card.FaceUp {
			if isSelected {
				cards = append(cards, selectedStyle.Render(formatCard(card)))
			} else {
				cards = append(cards, formatCard(card))
			}
		} else {
			cards = append(cards, cardStyle.Render("###"))
		}
	}

	// Build column from the bottom up, showing only cards that fit in maxHeight
	// Each card contributes 1 visible line (4 height - 3 overlap = 1 new line per card)
	// First card contributes 4 lines
	visibleLines := 4 + (len(cards)-1)*1
	if visibleLines > maxHeight && len(cards) > 1 {
		// Calculate how many cards we can show
		// maxHeight = 4 + (n-1)*1  =>  n = maxHeight - 3
		maxCards := maxHeight - 3
		if maxCards < 1 {
			maxCards = 1
		}
		if maxCards < len(cards) {
			// Skip cards from the bottom (show top cards)
			cards = cards[len(cards)-maxCards:]
		}
	}

	// Join cards vertically with overlap
	col := header + "\n"
	for i, c := range cards {
		if i == 0 {
			col += c
		} else {
			col = overlap(col, c)
		}
	}

	return col
}

// overlap joins two multi-line strings so the second overwrites the last 3 lines of the first
func overlap(base, overlay string) string {
	baseLines := splitLines(base)
	overlayLines := splitLines(overlay)

	// Keep all but last 3 lines of base, then append overlay
	keep := len(baseLines) - 3
	if keep < 0 {
		keep = 0
	}
	result := append(baseLines[:keep], overlayLines...)
	return joinLines(result)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func formatCard(c engine.Card) string {
	symbol := c.Suit.Symbol() + "\uFE0F" // emoji variation selector for larger display
	rank := c.Rank.String()
	style := redCardStyle
	if !c.Suit.IsRed() {
		style = blackCardStyle
	}
	return style.Render(rank + symbol)
}

func Run() error {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
