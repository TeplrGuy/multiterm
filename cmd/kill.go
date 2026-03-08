package cmd

import (
	"fmt"

	"github.com/TeplrGuy/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var killAll bool

var killCmd = &cobra.Command{
	Use:   "kill [session-name]",
	Short: "Kill a multiterm session",
	Long: `Kill a multiterm session by name, or all sessions with --all.

Examples:
  multiterm kill mt-1709901234    Kill a specific session
  multiterm kill --all            Kill all multiterm sessions`,
	Args: cobra.MaximumNArgs(1),
	RunE: runKill,
}

func init() {
	killCmd.Flags().BoolVarP(&killAll, "all", "a", false, "kill all multiterm sessions")
}

func runKill(cmd *cobra.Command, args []string) error {
	if killAll {
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No active multiterm sessions.")
			return nil
		}
		for _, s := range sessions {
			if err := tmux.KillSession(s); err != nil {
				fmt.Printf("  ✗ %s: %v\n", s, err)
			} else {
				fmt.Printf("  ✓ killed %s\n", s)
			}
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("specify a session name or use --all")
	}

	name := args[0]
	if !tmux.SessionExists(name) {
		return fmt.Errorf("session %q not found", name)
	}

	if err := tmux.KillSession(name); err != nil {
		return err
	}
	fmt.Printf("✓ killed %s\n", name)
	return nil
}
