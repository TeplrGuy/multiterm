# multiterm

> Open multiple terminal panes instantly with smart layouts. Powered by tmux.

[![CI](https://github.com/TeplrGuy/multiterm/actions/workflows/ci.yml/badge.svg)](https://github.com/TeplrGuy/multiterm/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/TeplrGuy/multiterm)](https://github.com/TeplrGuy/multiterm/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/TeplrGuy/multiterm)](https://goreportcard.com/report/github.com/TeplrGuy/multiterm)

```
┌──────────┬──────────┬──────────┐
│          │          │          │
│   api    │   logs   │    db    │
│          │          │          │
├──────────┼──────────┼──────────┤
│          │          │          │
│  tests   │  shell   │  htop   │
│          │          │          │
└──────────┴──────────┴──────────┘
```

**One command. Multiple panes. Named. Interactive. Zero friction.**

## Features

- 🖥️ **Configurable pane count** — 1 to 20 panes in a single command
- 📐 **Smart layouts** — grid, vertical, horizontal, main-side
- 🏷️ **Named panes** — label each pane (`-c "api:npm start"`)
- 📋 **Profiles** — save and reuse workspace setups
- 🔄 **Broadcast mode** — type in all panes simultaneously (`--sync`)
- 🌐 **SSH multi-host** — one pane per server (`--hosts s1,s2,s3`)
- 🤖 **Copilot profile** — built-in GitHub Copilot CLI integration
- 🖱️ **Mouse support** — click any pane to focus and type
- 🔧 **Auto-install** — installs tmux via Homebrew if missing
- 💾 **Session save** — capture your layout as a reusable profile

## Install

### Homebrew (macOS & Linux)

```bash
brew install TeplrGuy/tap/multiterm
```

### Go Install

```bash
go install github.com/TeplrGuy/multiterm@latest
```

### Download Binary

Download the latest binary from the [Releases page](https://github.com/TeplrGuy/multiterm/releases) and place it in your PATH:

```bash
# After downloading the archive for your platform:
tar xzf multiterm_*.tar.gz
sudo mv multiterm /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/TeplrGuy/multiterm.git
cd multiterm
make install
```

## Quick Start

```bash
# Open 6 panes in a grid (default)
multiterm

# Open 4 panes
multiterm 4

# Named panes with commands
multiterm -c "api:npm start" -c "logs:tail -f app.log" -c "db:psql"

# Vertical stack layout
multiterm -l vertical

# Broadcast to all panes (great for multi-server ops)
multiterm --sync

# SSH into multiple servers at once
multiterm --hosts user@server1,user@server2,user@server3

# Launch with GitHub Copilot CLI in the main pane
multiterm -p copilot
```

## Usage

```
multiterm [pane-count] [flags]
multiterm [command]

Commands:
  init        Create a default ~/.multiterm.yaml config file
  save        Save current session as a reusable profile
  list        List active multiterm sessions
  kill        Kill a multiterm session

Flags:
  -n, --count int         Number of panes (default: 6)
  -l, --layout string     Layout: grid, vertical, horizontal, main-side
  -p, --profile string    Use a saved profile from config
  -c, --cmd stringArray   Command to run in a pane (name:cmd or cmd)
      --sync              Broadcast input to all panes simultaneously
      --hosts string      Comma-separated SSH hosts (one pane per host)
  -h, --help              Help for multiterm
  -v, --version           Version
```

## Named Panes

Give each pane a descriptive label that shows in the border:

```bash
# Syntax: name:command
multiterm -c "api:npm start" -c "logs:tail -f app.log" -c "tests:npm test"
```

```
┌─ api ──────────┬─ logs ─────────┬─ tests ────────┐
│ npm start      │ [app.log]      │ npm test       │
│ Server on :3000│ GET /api/users │ 47 passing     │
└────────────────┴────────────────┴────────────────┘
```

Without a name prefix, panes are labeled `shell-1`, `shell-2`, etc.

## Broadcast Mode

Type the same input into all panes simultaneously:

```bash
# Start with sync enabled
multiterm --sync

# Or toggle inside a session with Ctrl-b B
```

Perfect for running the same command across multiple servers or repos.

## SSH Multi-Host

Open one pane per SSH host, with hostname labels:

```bash
multiterm --hosts user@web1,user@web2,user@db1

# Combine with --sync for cluster management
multiterm --hosts user@web1,user@web2,user@web3 --sync
```

## Copilot Profile

Launch GitHub Copilot CLI alongside your terminal shells:

```bash
multiterm -p copilot
```

```
┌─ copilot ──────────────┬─ shell-1 ──────┐
│                        │                 │
│  GitHub Copilot CLI    ├─ shell-2 ──────┤
│  Ask me anything...    │                 │
│                        │                 │
└────────────────────────┴─────────────────┘
```

Main pane (60%) runs `copilot`, side panes are interactive shells.

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
      - "editor:nvim ."
      - "server:npm run dev"
      - "logs:tail -f logs/app.log"
      - ""

  monitor:
    count: 6
    layout: grid
    commands:
      - "cpu:htop"
      - "io:iostat 1"
      - "docker:docker stats"
      - "syslog:tail -f /var/log/system.log"
      - "disk:watch df -h"
      - ""
```

Then use a profile:

```bash
multiterm -p dev
multiterm -p monitor
```

### Save a Session as a Profile

Set up your panes the way you like, then save it:

```bash
multiterm save my-workflow
```

This captures the current layout and pane labels into `~/.multiterm.yaml` for reuse.

## Session Management

```bash
# List all active multiterm sessions
multiterm list

# Kill a specific session
multiterm kill mt-1709901234

# Kill all multiterm sessions
multiterm kill --all
```

## Keyboard Shortcuts

Inside a multiterm session, standard tmux keybindings work:

| Shortcut | Action |
|----------|--------|
| **Click pane** | Switch focus to that pane |
| `Ctrl-b B` | Toggle broadcast mode (sync input to all panes) |
| `Ctrl-b d` | Detach from session (keeps it running) |
| `Ctrl-b arrow` | Navigate between panes |
| `Ctrl-b z` | Zoom/unzoom current pane (fullscreen) |
| `Ctrl-b x` | Close current pane |

## Environment Variables

Each pane gets environment variables for scripting and tool integration:

| Variable | Description |
|----------|-------------|
| `MULTITERM_SESSION` | Session name (e.g., `mt-1709901234`) |
| `MULTITERM_PANE_ID` | Pane ID (e.g., `%5`) |
| `MULTITERM_PANE_NAME` | Pane label (e.g., `api`, `shell-1`) |

## How It Works

multiterm is an intelligent wrapper around [tmux](https://github.com/tmux/tmux). It:

1. Ensures tmux is installed (auto-installs via Homebrew on macOS)
2. Creates a detached tmux session with mouse support
3. Splits panes according to the chosen layout
4. Labels each pane and sets environment variables
5. Sends specified commands to each pane
6. Attaches you to the session

## Requirements

- **tmux** — auto-installed via Homebrew on macOS, or `sudo apt install tmux` on Linux
- Works in any terminal emulator (Terminal.app, iTerm2, Alacritty, Kitty, Warp, etc.)

## CI/CD

- **Every push and PR** is tested (build, lint, vet) on macOS and Linux
- **Dependency updates** are checked weekly via Dependabot
- **Releases** are automated — push a tag (`git tag v2.0.0 && git push --tags`) and goreleaser builds binaries for all platforms

## Contributing

Contributions welcome! Please open an issue or submit a PR.

```bash
git clone https://github.com/TeplrGuy/multiterm.git
cd multiterm
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
```

## License

[MIT](LICENSE) © Gilbert Appiah
