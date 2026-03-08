package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up multiterm for all Copilot CLI sessions",
	Long: `One-time setup that makes multiterm available in every Copilot CLI session.

This command:
  1. Registers multiterm as a global MCP server in ~/.copilot/mcp-config.json
  2. Installs the multiterm skill to ~/.copilot/skills/multiterm/
  3. Installs the multiterm agent to ~/.copilot/agents/

After setup, every Copilot CLI session automatically has access to
multiterm's pane management tools. Copilot will use visible panes
for builds, tests, and agent work — no special launch command needed.

Safe to run multiple times (idempotent).`,
	SilenceUsage: true,
	RunE:         runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}

	copilotDir := filepath.Join(homeDir, ".copilot")

	// Validate multiterm is in PATH
	multitermPath, err := exec.LookPath("multiterm")
	if err != nil {
		fmt.Println("⚠ multiterm not found in PATH — MCP config will use 'multiterm' (install it first)")
		multitermPath = "multiterm"
	}

	var steps int

	// Step 1: Register global MCP server
	if err := setupMCPConfig(copilotDir, multitermPath); err != nil {
		return fmt.Errorf("MCP config: %w", err)
	}
	steps++
	fmt.Println("  ✓ Registered multiterm MCP server in ~/.copilot/mcp-config.json")

	// Step 2: Install skill
	if err := setupSkill(copilotDir); err != nil {
		return fmt.Errorf("skill install: %w", err)
	}
	steps++
	fmt.Println("  ✓ Installed multiterm skill to ~/.copilot/skills/multiterm/")

	// Step 3: Install agent
	if err := setupAgent(copilotDir); err != nil {
		return fmt.Errorf("agent install: %w", err)
	}
	steps++
	fmt.Println("  ✓ Installed multiterm agent to ~/.copilot/agents/")

	fmt.Printf("\n✦ Setup complete (%d components installed)\n", steps)
	fmt.Println("  Every Copilot CLI session now has multiterm pane tools.")
	fmt.Println("  When Copilot runs tasks, they'll appear as visible terminal panes.")
	fmt.Println("\n  To watch: tmux attach -t mt-copilot")

	return nil
}

// setupMCPConfig adds multiterm to ~/.copilot/mcp-config.json
func setupMCPConfig(copilotDir, multitermPath string) error {
	configPath := filepath.Join(copilotDir, "mcp-config.json")

	// Read existing config or start fresh
	existing := make(map[string]any)
	data, err := os.ReadFile(configPath)
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &existing); err != nil {
			// Backup corrupted file
			backupPath := configPath + ".bak"
			os.WriteFile(backupPath, data, 0644)
			fmt.Printf("  ⚠ Backed up malformed config to %s\n", backupPath)
			existing = make(map[string]any)
		}
	}

	// Add/update multiterm entry
	existing["multiterm"] = map[string]any{
		"type":    "stdio",
		"command": multitermPath,
		"args":    []string{"serve"},
	}

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, out, 0644)
}

// setupSkill installs the multiterm skill to ~/.copilot/skills/multiterm/
func setupSkill(copilotDir string) error {
	skillDir := filepath.Join(copilotDir, "skills", "multiterm")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	return os.WriteFile(skillPath, []byte(skillContent), 0644)
}

// setupAgent installs the multiterm agent to ~/.copilot/agents/
func setupAgent(copilotDir string) error {
	agentsDir := filepath.Join(copilotDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return err
	}

	agentPath := filepath.Join(agentsDir, "multiterm.agent.md")
	return os.WriteFile(agentPath, []byte(agentContent), 0644)
}

const skillContent = `---
name: multiterm
description: 'Make agent work visible in terminal panes. Use when running tests, builds, deployments, background tasks, monitoring, agents, sub-agents, or any command the user should see. Triggers on: run, test, build, deploy, agent, background, monitor, watch, parallel, visible, pane.'
license: MIT
allowed-tools: Bash
---

# multiterm — Visible Terminal Panes for Agent Work

## When to Use This Skill

Whenever you need to run commands that the user should see in real time:
- Running tests (go test, npm test, pytest, etc.)
- Building projects (make, go build, npm run build)
- Starting servers or services
- Tailing logs
- Running multiple parallel tasks
- Any background or sub-agent work

## How It Works

multiterm is registered as a global MCP server. You have 6 tools available:

| Tool | Use For |
|------|---------|
| create_pane | Open a new labeled terminal pane |
| run_in_pane | Run a command in an existing pane |
| read_pane | Capture output to check results |
| list_panes | See all active panes |
| close_pane | Clean up finished panes |
| broadcast | Send command to all panes |

## Workflow Patterns

### Running Tests
` + "```" + `
1. create_pane(name="tests", command="go test ./... -v")
2. (wait for completion)
3. read_pane(pane_id="<id>", lines=30)
4. Report results to user
` + "```" + `

### Build + Test in Parallel
` + "```" + `
1. create_pane(name="build", command="make build")
2. create_pane(name="lint", command="make lint")
3. read_pane each to check results
` + "```" + `

### Log Monitoring
` + "```" + `
1. create_pane(name="logs", command="tail -f /var/log/app.log")
2. (leave running for user to watch)
` + "```" + `

### Multi-Step Deployment
` + "```" + `
1. create_pane(name="deploy-1", command="make build")
2. read_pane → verify success
3. run_in_pane(pane_id, "make deploy")
4. read_pane → verify deployment
` + "```" + `

## Key Rules

1. **Always create a pane** before running significant commands
2. **Label panes clearly** — "tests", "build", "logs", not "pane-1"
3. **Read output** after commands to verify success/failure
4. **Leave service panes open** — only close on-demand task panes
5. **Tell the user** to run ` + "`tmux attach -t mt-copilot`" + ` if they want to watch

## Auto-Session

If no multiterm session exists, the MCP server auto-creates one called
` + "`mt-copilot`" + `. The user can view it anytime with:

` + "```" + `bash
tmux attach -t mt-copilot
` + "```" + `
`
