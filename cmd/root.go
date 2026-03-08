package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gilbertappiah/multiterm/internal/config"
	"github.com/gilbertappiah/multiterm/internal/layout"
	"github.com/gilbertappiah/multiterm/internal/setup"
	"github.com/gilbertappiah/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var (
	flagCount   int
	flagLayout  string
	flagProfile string
	flagCmds    []string
	version     = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "multiterm [pane-count]",
	Short: "Open multiple terminal panes instantly",
	Long: `multiterm — a fast terminal multiplexer powered by tmux.

Open a configurable number of terminal panes in a single window
with smart layouts and profile support.

Examples:
  multiterm              Open 6 panes in a grid (default)
  multiterm 4            Open 4 panes in a grid
  multiterm -l vertical  Open panes stacked vertically
  multiterm -p dev       Use the "dev" profile from ~/.multiterm.yaml
  multiterm -c "htop" -c "npm start"  Run commands in panes`,
	Version:      version,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&flagCount, "count", "n", 0, "number of panes to open")
	rootCmd.Flags().StringVarP(&flagLayout, "layout", "l", "", "layout: grid, vertical, horizontal, main-side")
	rootCmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "use a named profile from ~/.multiterm.yaml")
	rootCmd.Flags().StringArrayVarP(&flagCmds, "cmd", "c", nil, "command to run in a pane (repeatable)")

	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(initConfigCmd)
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Ensure tmux is available.
	if err := setup.EnsureTmux(); err != nil {
		return err
	}

	// Load config.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Determine pane count and layout.
	paneCount := cfg.Defaults.Count
	layoutName := cfg.Defaults.Layout
	var commands []string

	// Positional arg overrides count.
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 20 {
			return fmt.Errorf("invalid pane count %q: must be a number between 1 and 20", args[0])
		}
		paneCount = n
	}

	// Profile overrides everything.
	if flagProfile != "" {
		profile, err := config.GetProfile(cfg, flagProfile)
		if err != nil {
			return err
		}
		paneCount = profile.Count
		if profile.Layout != "" {
			layoutName = profile.Layout
		}
		commands = profile.Commands
	}

	// Explicit flags override profile/defaults.
	if flagCount > 0 {
		paneCount = flagCount
	}
	if flagLayout != "" {
		layoutName = flagLayout
	}
	if len(flagCmds) > 0 {
		commands = flagCmds
	}

	// Validate layout.
	layoutName, err = layout.ParseLayout(layoutName)
	if err != nil {
		return err
	}

	// Calculate layout plan.
	plan, err := layout.Calculate(layoutName, paneCount)
	if err != nil {
		return err
	}

	// Generate session name.
	sessionName := fmt.Sprintf("mt-%d", time.Now().Unix())

	// Create tmux session.
	if err := tmux.NewSession(sessionName); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Enable mouse mode, pane borders, and status bar.
	tmux.ConfigureSession(sessionName)

	// Get the first pane's ID.
	firstPane, err := tmux.FirstPaneID(sessionName)
	if err != nil {
		_ = tmux.KillSession(sessionName)
		return fmt.Errorf("failed to get first pane: %w", err)
	}

	// Create panes according to the layout plan.
	paneIDs := []string{firstPane}
	for _, split := range plan.Splits {
		targetIdx := split.TargetPane
		if targetIdx >= len(paneIDs) {
			targetIdx = len(paneIDs) - 1
		}
		target := paneIDs[targetIdx]

		var newPaneID string
		switch split.Direction {
		case layout.SplitH:
			newPaneID, err = tmux.SplitVertical(sessionName, target)
		case layout.SplitV:
			newPaneID, err = tmux.SplitHorizontal(sessionName, target)
		}
		if err != nil {
			_ = tmux.KillSession(sessionName)
			return fmt.Errorf("failed to create pane: %w", err)
		}
		paneIDs = append(paneIDs, newPaneID)
	}

	// Apply the tmux built-in layout.
	if plan.TmuxLayoutName != "" {
		if err := tmux.SelectLayout(sessionName, plan.TmuxLayoutName); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not apply layout: %v\n", err)
		}
	}

	// Send commands to panes.
	for i, c := range commands {
		if i >= len(paneIDs) {
			break
		}
		if c != "" {
			if err := tmux.SendCommand(sessionName, paneIDs[i], c); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to send command to pane %d: %v\n", i, err)
			}
		}
	}

	// Print info and attach.
	fmt.Printf("✦ multiterm — %d panes [%s] session: %s\n", paneCount, layoutName, sessionName)
	fmt.Println("  Click any pane to switch focus. Ctrl-b d to detach.")

	// Focus the first pane.
	_ = tmux.SelectPane(sessionName, paneIDs[0])

	return tmux.AttachSession(sessionName)
}
