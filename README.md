# multiterm

> Open multiple terminal panes instantly with smart layouts. Powered by tmux.

[![Release](https://img.shields.io/github/v/release/gilbertappiah/multiterm)](https://github.com/gilbertappiah/multiterm/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gilbertappiah/multiterm)](https://goreportcard.com/report/github.com/gilbertappiah/multiterm)

```
┌──────────┬──────────┬──────────┐
│          │          │          │
│  pane 1  │  pane 2  │  pane 3  │
│          │          │          │
├──────────┼──────────┼──────────┤
│          │          │          │
│  pane 4  │  pane 5  │  pane 6  │
│          │          │          │
└──────────┴──────────┴──────────┘
```

**One command. Multiple panes. Zero friction.**

## Install

### Homebrew (macOS & Linux)

```bash
brew install gilbertappiah/tap/multiterm
```

### Go Install

```bash
go install github.com/gilbertappiah/multiterm@latest
```

### Download Binary

Download the latest binary from the [Releases page](https://github.com/gilbertappiah/multiterm/releases) and place it in your PATH:

```bash
# After downloading the archive for your platform:
tar xzf multiterm_*.tar.gz
sudo mv multiterm /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/gilbertappiah/multiterm.git
cd multiterm
make install
```

## Quick Start

```bash
# Open 6 panes in a grid (default)
multiterm

# Open 4 panes
multiterm 4

# Vertical stack layout
multiterm -l vertical

# Run commands in each pane
multiterm -c "htop" -c "npm run dev" -c "tail -f app.log"
```

## Usage

```
multiterm [pane-count] [flags]
multiterm [command]

Commands:
  init        Create a default ~/.multiterm.yaml config file
  kill        Kill a multiterm session
  list        List active multiterm sessions

Flags:
  -n, --count int         Number of panes (default: 6)
  -l, --layout string     Layout: grid, vertical, horizontal, main-side
  -p, --profile string    Use a saved profile from config
  -c, --cmd stringArray   Command to run in a pane (repeatable)
  -h, --help              Help for multiterm
  -v, --version           Version
```

## Layouts

| Layout | Flag | Description |
|--------|------|-------------|
| **Grid** | `-l grid` | Auto-calculated rows × columns (default) |
| **Vertical** | `-l vertical` | Panes stacked top-to-bottom |
| **Horizontal** | `-l horizontal` | Panes side-by-side |
| **Main + Side** | `-l main-side` | One large pane (60%) + stacked side panes |

### Grid (default)
```
┌────────┬────────┬────────┐
│   1    │   2    │   3    │
├────────┼────────┼────────┤
│   4    │   5    │   6    │
└────────┴────────┴────────┘
```

### Vertical
```
┌──────────────────────────┐
│            1             │
├──────────────────────────┤
│            2             │
├──────────────────────────┤
│            3             │
└──────────────────────────┘
```

### Horizontal
```
┌────────┬────────┬────────┐
│        │        │        │
│   1    │   2    │   3    │
│        │        │        │
└────────┴────────┴────────┘
```

### Main + Side
```
┌───────────────┬──────────┐
│               │    2     │
│               ├──────────┤
│       1       │    3     │
│               ├──────────┤
│               │    4     │
└───────────────┴──────────┘
```

## Config File

Generate a starter config:

```bash
multiterm init
```

This creates `~/.multiterm.yaml`:

```yaml
defaults:
  count: 6
  layout: grid

profiles:
  dev:
    count: 4
    layout: main-side
    commands:
      - "nvim ."
      - "npm run dev"
      - "tail -f logs/app.log"
      - ""

  monitor:
    count: 6
    layout: grid
    commands:
      - "htop"
      - "iostat 1"
      - "docker stats"
      - "tail -f /var/log/system.log"
      - "watch df -h"
      - ""
```

Then use a profile:

```bash
multiterm -p dev
multiterm -p monitor
```

## Session Management

```bash
# List all active multiterm sessions
multiterm list

# Kill a specific session
multiterm kill mt-1709901234

# Kill all multiterm sessions
multiterm kill --all
```

## How It Works

multiterm is a thin, intelligent wrapper around [tmux](https://github.com/tmux/tmux). It:

1. Ensures tmux is installed (auto-installs via Homebrew on macOS)
2. Creates a detached tmux session
3. Splits panes according to the chosen layout
4. Applies tmux's built-in layout engine for pixel-perfect arrangement
5. Sends any specified commands to each pane
6. Attaches you to the session

**Tmux keybindings still work** — `Ctrl-b` is your prefix key. Use `Ctrl-b d` to detach, `Ctrl-b arrow` to navigate panes.

## Requirements

- **tmux** — auto-installed via Homebrew on macOS, or `sudo apt install tmux` on Linux
- Works in any terminal emulator (Terminal.app, iTerm2, Alacritty, Kitty, Warp, etc.)

## Contributing

Contributions welcome! Please open an issue or submit a PR.

```bash
git clone https://github.com/gilbertappiah/multiterm.git
cd multiterm
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
```

## License

[MIT](LICENSE) © Gilbert Appiah
