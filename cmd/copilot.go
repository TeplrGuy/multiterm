package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/TeplrGuy/multiterm/internal/layout"
	"github.com/TeplrGuy/multiterm/internal/setup"
	"github.com/TeplrGuy/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var (
	copilotPanes int
	copilotModel string
)

var copilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Launch Copilot CLI with full pane integration",
	Long: `Launch GitHub Copilot CLI inside a multiterm session with the MCP
server pre-configured. Copilot gets access to pane management tools
so all agent work is visible in your terminal.

  Main pane:  Copilot CLI (with multiterm MCP tools enabled)
  Side panes: Interactive shells for your work

When Copilot runs sub-tasks, they can appear as visible panes instead
of invisible background processes.

Examples:
  multiterm copilot                  Launch with default 2 side panes
  multiterm copilot --panes 4        Launch with 4 side panes
  multiterm copilot --model gpt-4.1  Use a specific model`,
	SilenceUsage: true,
	RunE:         runCopilotCmd,
}

func init() {
	copilotCmd.Flags().IntVar(&copilotPanes, "panes", 2, "number of side panes")
	copilotCmd.Flags().StringVar(&copilotModel, "model", "", "model to pass to copilot CLI")
	rootCmd.AddCommand(copilotCmd)
}

// mcpConfig generates the JSON config for registering multiterm as an MCP server.
func mcpConfig(session string) (string, error) {
	multitermBin, err := os.Executable()
	if err != nil {
		multitermBin = "multiterm"
	}

	config := map[string]any{
		"multiterm": map[string]any{
			"type":    "stdio",
			"command": multitermBin,
			"args":    []string{"serve"},
		},
	}

	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func runCopilotCmd(cmd *cobra.Command, args []string) error {
	if err := setup.EnsureTmux(); err != nil {
		return err
	}

	// Find copilot CLI.
	copilotBin, err := exec.LookPath("copilot")
	if err != nil {
		return fmt.Errorf("copilot CLI not found in PATH — install it from https://github.com/github/copilot-cli")
	}

	totalPanes := 1 + copilotPanes // main + side panes
	layoutName := "main-vertical"
	if totalPanes <= 2 {
		layoutName = "even-horizontal"
	}

	plan, err := layout.Calculate("main-side", totalPanes)
	if err != nil {
		return err
	}

	sessionName := fmt.Sprintf("mt-%d", time.Now().Unix())

	paneIDs, err := createSession(sessionName, plan)
	if err != nil {
		return err
	}

	// Label all panes.
	labels := make([]string, len(paneIDs))
	labels[0] = "copilot"
	for i := 1; i < len(labels); i++ {
		labels[i] = fmt.Sprintf("shell-%d", i)
	}

	for i, paneID := range paneIDs {
		_ = tmux.RenamePane(sessionName, paneID, labels[i])
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_SESSION", sessionName)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_ID", paneID)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_NAME", labels[i])
		_ = tmux.SendCommand(sessionName, paneID, "clear")
	}

	// Generate MCP config and write to temp file.
	mcpJSON, err := mcpConfig(sessionName)
	if err != nil {
		return fmt.Errorf("failed to generate MCP config: %w", err)
	}

	tmpDir := os.TempDir()
	mcpConfigPath := filepath.Join(tmpDir, fmt.Sprintf("multiterm-mcp-%s.json", sessionName))
	if err := os.WriteFile(mcpConfigPath, []byte(mcpJSON), 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	// Build copilot CLI command with MCP integration.
	copilotArgs := []string{copilotBin}
	copilotArgs = append(copilotArgs, "--additional-mcp-config", mcpConfigPath)

	if copilotModel != "" {
		copilotArgs = append(copilotArgs, "--model", copilotModel)
	}

	// Launch copilot in the main pane.
	copilotCommand := strings.Join(copilotArgs, " ")
	_ = tmux.SendCommand(sessionName, paneIDs[0], copilotCommand)

	// Apply main-vertical layout for the copilot profile.
	_ = tmux.SelectLayout(sessionName, layoutName)

	fmt.Printf("✦ multiterm copilot — %d panes [main-side] session: %s\n", len(paneIDs), sessionName)
	fmt.Println("  Copilot has multiterm MCP tools │ Ctrl-b A: add pane │ Ctrl-b d: detach")
	fmt.Printf("  MCP config: %s\n", mcpConfigPath)

	_ = tmux.SelectPane(sessionName, paneIDs[0])
	return tmux.AttachSession(sessionName)
}
