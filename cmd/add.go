package cmd

import (
	"fmt"
	"strings"

	"github.com/TeplrGuy/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [session]",
	Short: "Add a new pane to a running multiterm session",
	Long: `Add a new pane to an existing multiterm session.

If no session name is given, the most recently created mt-* session is used.

Examples:
  multiterm add                         Add a shell pane to the latest session
  multiterm add -c "logs:tail -f app.log"  Add a named pane with a command
  multiterm add mt-1234567890           Add a pane to a specific session
  multiterm add --vertical              Split vertically instead of auto-tiling`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

var (
	addFlagCmd      string
	addFlagVertical bool
)

func init() {
	addCmd.Flags().StringVarP(&addFlagCmd, "cmd", "c", "", "command to run (name:cmd or cmd)")
	addCmd.Flags().BoolVarP(&addFlagVertical, "vertical", "v", false, "split vertically (side-by-side) instead of auto-tiling")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	sessionName := ""

	if len(args) == 1 {
		sessionName = args[0]
	} else {
		// Find the most recent mt-* session.
		sessions, err := tmux.ListSessions()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		if len(sessions) == 0 {
			return fmt.Errorf("no multiterm sessions running — start one with: multiterm")
		}
		// Sessions are listed chronologically; take the last one.
		sessionName = sessions[len(sessions)-1]
	}

	if !tmux.SessionExists(sessionName) {
		return fmt.Errorf("session %q not found", sessionName)
	}

	// Get current pane count for labeling.
	panes, err := tmux.ListPanes(sessionName)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}
	nextIndex := len(panes) + 1

	// Parse the command spec.
	label := fmt.Sprintf("shell-%d", nextIndex)
	command := ""
	if addFlagCmd != "" {
		spec := parseCommand(addFlagCmd)
		if spec.Name != "" {
			label = spec.Name
		}
		command = spec.Command
	}

	// Split from the currently active pane.
	activePaneID := ""
	for _, p := range panes {
		activePaneID = p.ID // will use last; tmux active pane is selected anyway
	}
	if activePaneID == "" {
		return fmt.Errorf("no panes found in session %s", sessionName)
	}

	var newPaneID string
	if addFlagVertical {
		newPaneID, err = tmux.SplitVertical(sessionName, activePaneID)
	} else {
		newPaneID, err = tmux.SplitHorizontal(sessionName, activePaneID)
	}
	if err != nil {
		return fmt.Errorf("failed to split pane: %w", err)
	}

	// Label the new pane.
	_ = tmux.RenamePane(sessionName, newPaneID, label)
	_ = tmux.SetEnv(sessionName, newPaneID, "MULTITERM_SESSION", sessionName)
	_ = tmux.SetEnv(sessionName, newPaneID, "MULTITERM_PANE_ID", newPaneID)
	_ = tmux.SetEnv(sessionName, newPaneID, "MULTITERM_PANE_NAME", label)
	_ = tmux.SendCommand(sessionName, newPaneID, "clear")

	if command != "" {
		_ = tmux.SendCommand(sessionName, newPaneID, command)
	}

	// Re-tile the layout so all panes are evenly distributed.
	_ = tmux.SelectLayout(sessionName, "tiled")

	// Show the pane border with updated title.
	parts := []string{label}
	if command != "" {
		parts = append(parts, command)
	}
	fmt.Printf("✦ Added pane [%s] to %s (%d total)\n", strings.Join(parts, ": "), sessionName, nextIndex)

	return nil
}
