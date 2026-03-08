---
name: multiterm
description: 'Terminal pane manager for visible agent work. When inside a multiterm session, use MCP tools (create_pane, run_in_pane, read_pane) to make all work visible instead of running invisibly in the background.'
tools: Bash
---

# multiterm — Visible Agent Work

You are inside a multiterm terminal session. You have access to MCP tools that let you
manage terminal panes. **Always prefer visible panes over invisible background work.**

## Core Principle

When you need to run commands, tests, builds, or any sub-task — open a **visible pane**
so the user can see what's happening in real time. The user hates invisible background
processes.

## Available MCP Tools

### `create_pane`
Open a new terminal pane with a label and optional command.
```
name: "test-runner"
command: "npm test"
```

### `run_in_pane`
Execute a command in an existing pane.
```
pane_id: "%5"
command: "go build ./..."
```

### `read_pane`
Capture output from a pane to see results.
```
pane_id: "%5"
lines: 50
```

### `list_panes`
See all active panes with their IDs and names.

### `close_pane`
Clean up a pane when its task is done.
```
pane_id: "%5"
```

### `broadcast`
Send the same command to every pane at once.
```
command: "git pull"
```

## Workflow

1. **Before running a task**: Create a named pane for it
2. **Run the command** in that pane
3. **Read the output** to check results
4. **Close the pane** when done (or leave it for the user)

## Example: Running Tests

```
1. create_pane(name="tests", command="npm test")
2. Wait a moment for tests to run
3. read_pane(pane_id=<id from step 1>, lines=30)
4. Analyze results and report to user
```

## Example: Multi-Service Setup

```
1. create_pane(name="api", command="npm run dev")
2. create_pane(name="frontend", command="cd frontend && npm start")
3. create_pane(name="db", command="docker compose up db")
4. list_panes() to confirm everything is running
```

## Guidelines

- **Label panes clearly** — use descriptive names like "test-runner", "build", "logs"
- **Read output** after commands complete to verify success
- **Don't close panes** with long-running services unless asked
- **Use broadcast** for operations that should hit all panes (like git pull)
- **Check MULTITERM_SESSION env var** — if it's set, you're inside multiterm
