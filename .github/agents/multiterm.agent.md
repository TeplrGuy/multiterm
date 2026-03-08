---
name: multiterm
description: 'Terminal pane manager for visible agent work. Use MCP tools (create_pane, run_in_pane, read_pane) to make all work visible instead of running invisibly in the background. Works from ANY Copilot CLI session after running multiterm setup.'
tools: Bash
---

# multiterm — Visible Agent Work

You have access to MCP tools that manage terminal panes via multiterm.
**Always prefer visible panes over invisible background work.**

If no multiterm session exists, the tools will auto-create one. Tell the
user to run `tmux attach -t mt-copilot` in another terminal to watch.

## Core Principle

When you need to run commands, tests, builds, or any long-running task — open a
**visible pane** so the user can see what is happening in real time.

## Available MCP Tools

| Tool | Purpose |
|------|---------|
| `create_pane` | Open a labeled pane with an optional command |
| `run_in_pane` | Execute a command in an existing pane |
| `read_pane` | Capture output from a pane (default 50 lines) |
| `list_panes` | List all panes with IDs, names, and commands |
| `close_pane` | Clean up a pane when its task is done |
| `broadcast` | Send the same command to every pane |

## When to Use Panes

- **Running tests**: `create_pane("tests", "go test ./...")` then `read_pane`
- **Building**: `create_pane("build", "make build")` then `read_pane`
- **Log tailing**: `create_pane("logs", "tail -f app.log")`
- **Multiple commands**: one pane per task so user sees everything
- **Sub-agent work**: always visible, never invisible

## Workflow

1. `list_panes` — check what is already running
2. `create_pane` — open a pane for the new task
3. `run_in_pane` — execute in an existing pane if reusing
4. `read_pane` — capture output to verify results
5. `close_pane` — clean up when done (optional)

## Guidelines

- Always label panes clearly (e.g. "tests", "build", "deploy")
- Read output after commands complete to verify success
- Do not close panes running long-lived services unless asked
- Use broadcast for operations that apply to all panes
