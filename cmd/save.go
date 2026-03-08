package cmd

import (
	"fmt"

	"github.com/gilbertappiah/multiterm/internal/config"
	"github.com/gilbertappiah/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save <profile-name> [session-name]",
	Short: "Save current session as a reusable profile",
	Long: `Save the current multiterm session's layout and pane configuration
as a named profile in ~/.multiterm.yaml.

If no session name is given, uses the most recent mt-* session.

Examples:
  multiterm save dev-setup
  multiterm save my-workflow mt-1709901234`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSave,
}

func runSave(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	// Find the session to save.
	var sessionName string
	if len(args) == 2 {
		sessionName = args[1]
	} else {
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			return fmt.Errorf("no active multiterm sessions to save")
		}
		sessionName = sessions[len(sessions)-1]
	}

	if !tmux.SessionExists(sessionName) {
		return fmt.Errorf("session %q not found", sessionName)
	}

	// Get pane info.
	panes, err := tmux.ListPanes(sessionName)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	// Build commands list from pane titles.
	commands := make([]string, len(panes))
	for i, p := range panes {
		if p.Title != "" && p.Title != "zsh" && p.Title != "bash" {
			commands[i] = fmt.Sprintf("%s:", p.Title)
		} else {
			commands[i] = ""
		}
	}

	// Load existing config and add the profile.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	layoutName := "grid"
	if len(panes) <= 3 {
		layoutName = "vertical"
	}

	profile := config.Profile{
		Count:    len(panes),
		Layout:   layoutName,
		Commands: commands,
	}

	if err := config.SaveProfile(cfg, profileName, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Printf("✓ Saved profile %q (%d panes, %s layout) to %s\n",
		profileName, len(panes), layoutName, config.ConfigPath())
	return nil
}
