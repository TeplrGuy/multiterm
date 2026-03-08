package cmd

import (
	"fmt"

	"github.com/gilbertappiah/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active multiterm sessions",
	Long:  "List all active tmux sessions created by multiterm.",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No active multiterm sessions.")
		return nil
	}

	fmt.Printf("Active multiterm sessions (%d):\n", len(sessions))
	for _, s := range sessions {
		fmt.Printf("  • %s\n", s)
	}
	return nil
}
