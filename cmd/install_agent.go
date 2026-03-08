package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installAgentCmd = &cobra.Command{
	Use:   "install-agent",
	Short: "Install the multiterm Copilot agent globally",
	Long: `Install the multiterm agent to ~/.copilot/agents/ so it's available
in all your Copilot CLI sessions.

The agent teaches Copilot to use multiterm's MCP tools for visible
pane management instead of invisible background processes.

After installing, start Copilot with:
  multiterm copilot`,
	SilenceUsage: true,
	RunE:         runInstallAgent,
}

func init() {
	rootCmd.AddCommand(installAgentCmd)
}

const agentContent = `---
name: multiterm
description: 'Terminal pane manager for visible agent work. Use MCP tools (create_pane, run_in_pane, read_pane) to make all work visible instead of running invisibly in the background. Works from ANY Copilot CLI session after running multiterm setup.'
tools: Bash
---

# multiterm — Visible Agent Work

You have access to MCP tools that manage terminal panes via multiterm.
**Always prefer visible panes over invisible background work.**

If no multiterm session exists, the tools will auto-create one. Tell the
user to run ` + "`tmux attach -t mt-copilot`" + ` in another terminal to watch.

## Core Principle

When you need to run commands, tests, builds, or any long-running task — open a
**visible pane** so the user can see what is happening in real time.

## Available MCP Tools

| Tool | Purpose |
|------|---------|
| create_pane | Open a labeled pane with an optional command |
| run_in_pane | Execute a command in an existing pane |
| read_pane | Capture output from a pane (default 50 lines) |
| list_panes | List all panes with IDs, names, and commands |
| close_pane | Clean up a pane when its task is done |
| broadcast | Send the same command to every pane |

## When to Use Panes

- **Running tests**: create_pane("tests", "go test ./...") then read_pane
- **Building**: create_pane("build", "make build") then read_pane
- **Log tailing**: create_pane("logs", "tail -f app.log")
- **Multiple commands**: one pane per task so user sees everything
- **Sub-agent work**: always visible, never invisible

## Workflow

1. ` + "`list_panes`" + ` — check what is already running
2. ` + "`create_pane`" + ` — open a pane for the new task
3. ` + "`run_in_pane`" + ` — execute in an existing pane if reusing
4. ` + "`read_pane`" + ` — capture output to verify results
5. ` + "`close_pane`" + ` — clean up when done (optional)

## Guidelines

- Always label panes clearly (e.g. "tests", "build", "deploy")
- Read output after commands complete to verify success
- Do not close panes running long-lived services unless asked
- Use broadcast for operations that apply to all panes
`

func runInstallAgent(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".copilot", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("could not create agents directory: %w", err)
	}

	agentPath := filepath.Join(agentsDir, "multiterm.agent.md")

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("could not write agent file: %w", err)
	}

	fmt.Printf("✦ Installed multiterm agent to %s\n", agentPath)
	fmt.Println("  The agent is now available in all Copilot CLI sessions.")

	if runtime.GOOS == "darwin" {
		fmt.Println("\n  Quick start:")
		fmt.Println("    multiterm copilot    Launch Copilot with pane integration")
	}

	return nil
}
