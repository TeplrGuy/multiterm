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

### create_pane
Open a new terminal pane with a label and optional command.
- name: descriptive label (e.g. "test-runner", "build", "logs")
- command: the command to run

### run_in_pane
Execute a command in an existing pane.
- pane_id: the pane ID from list_panes
- command: the command to execute

### read_pane
Capture output from a pane to see results.
- pane_id: the pane ID
- lines: number of lines to capture (default: 50)

### list_panes
See all active panes with their IDs, names, and running commands.

### close_pane
Clean up a pane when its task is done.
- pane_id: the pane ID to close

### broadcast
Send the same command to every pane at once.
- command: the command to broadcast

## Workflow

1. Before running a task: create a named pane for it
2. Run the command in that pane
3. Read the output to check results
4. Close the pane when done (or leave it for the user)

## Guidelines

- Label panes clearly with descriptive names
- Read output after commands complete to verify success
- Don't close panes with long-running services unless asked
- Use broadcast for operations that should hit all panes
- Check MULTITERM_SESSION env var to detect multiterm sessions
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
