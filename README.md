# Solitaire

A Klondike Solitaire game written in Go with both a web interface and a terminal UI (TUI).

## Features

- **Web Interface**: Play in your browser with HTMX for dynamic updates
- **Terminal UI**: Play directly in your terminal using Bubble Tea
- **Auto-move**: Automatically move eligible cards to foundation piles
- **Undo**: Undo your last move
- **Multiple draw modes**: Toggle between 1-card and 3-card draw
- **One-click move**: Click a card to automatically move it to a valid destination

## Installation

```bash
go get github.com/YOUR_USERNAME/solitaire
```

## Usage

### Web Mode

```bash
make run-web
# or
./solitaire web -a :8080
```

Then open http://localhost:8080 in your browser.

### TUI Mode

```bash
make run-tui
# or
./solitaire tui
```

#### TUI Controls

| Key | Action |
|-----|--------|
| `n` | New game |
| `d` | Draw from stock |
| `a` | Auto-move to foundation |
| `u` | Undo last move |
| `1-7` | Select/move to tableau pile |
| `w` | Select waste card |
| `f` | Move selected card to foundation |
| `space` | Clear selection |
| `?` | Toggle help |
| `q` | Quit |

## Development

```bash
# Build
make build

# Test
make test

# Clean
make clean
```

## Project Structure

```
.
├── cmd/solitaire/     # CLI entrypoint (Cobra)
├── engine/              # Core game engine (cards, piles, deck)
├── games/klondike/      # Klondike solitaire implementation
├── server/              # HTTP server and handlers
├── tui/                 # Terminal UI (Bubble Tea)
├── static/              # CSS and assets
└── server/templates/    # HTML templates
```

## License

MIT
