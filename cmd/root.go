package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TeplrGuy/multiterm/internal/config"
	"github.com/TeplrGuy/multiterm/internal/layout"
	"github.com/TeplrGuy/multiterm/internal/setup"
	"github.com/TeplrGuy/multiterm/internal/tmux"
	"github.com/spf13/cobra"
)

var (
	flagCount   int
	flagLayout  string
	flagProfile string
	flagCmds    []string
	flagSync    bool
	flagHosts   string
	version     = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "multiterm [pane-count]",
	Short: "Open multiple terminal panes instantly",
	Long: `multiterm — a fast terminal multiplexer powered by tmux.

Open a configurable number of terminal panes in a single window
with smart layouts and profile support.

Examples:
  multiterm                            Open 6 panes in a grid (default)
  multiterm 4                          Open 4 panes in a grid
  multiterm -l vertical                Open panes stacked vertically
  multiterm -p dev                     Use the "dev" profile
  multiterm -c "api:npm start" -c "logs:tail -f app.log"
                                       Named panes with commands
  multiterm --hosts user@s1,user@s2    One pane per SSH host
  multiterm --sync                     Broadcast input to all panes`,
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
	rootCmd.Flags().StringArrayVarP(&flagCmds, "cmd", "c", nil, "command to run in a pane (name:cmd or cmd)")
	rootCmd.Flags().BoolVar(&flagSync, "sync", false, "broadcast input to all panes simultaneously")
	rootCmd.Flags().StringVar(&flagHosts, "hosts", "", "comma-separated SSH hosts (one pane per host)")

	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(initConfigCmd)
	rootCmd.AddCommand(saveCmd)
}

// PaneSpec holds a parsed pane name and command.
type PaneSpec struct {
	Name    string
	Command string
}

// parseCommand parses "name:command" or just "command" syntax.
func parseCommand(raw string) PaneSpec {
	if raw == "" {
		return PaneSpec{}
	}
	// Look for the first colon that separates name from command.
	// Avoid splitting on colons inside commands (e.g., "http://...")
	idx := strings.Index(raw, ":")
	if idx > 0 && idx < len(raw)-1 && !strings.Contains(raw[:idx], "/") && !strings.Contains(raw[:idx], " ") {
		return PaneSpec{
			Name:    raw[:idx],
			Command: raw[idx+1:],
		}
	}
	return PaneSpec{Command: raw}
}

func runRoot(cmd *cobra.Command, args []string) error {
	if err := setup.EnsureTmux(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Handle SSH multi-host mode.
	if flagHosts != "" {
		return runSSHMultiHost(cfg)
	}

	paneCount := cfg.Defaults.Count
	layoutName := cfg.Defaults.Layout
	var commands []string

	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 20 {
			return fmt.Errorf("invalid pane count %q: must be a number between 1 and 20", args[0])
		}
		paneCount = n
	}

	// Profile overrides.
	if flagProfile != "" {
		// Check built-in profiles first.
		if flagProfile == "copilot" {
			return runCopilotProfile(cfg)
		}
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

	if flagCount > 0 {
		paneCount = flagCount
	}
	if flagLayout != "" {
		layoutName = flagLayout
	}
	if len(flagCmds) > 0 {
		commands = flagCmds
	}

	// Parse named pane specs.
	specs := make([]PaneSpec, len(commands))
	for i, c := range commands {
		specs[i] = parseCommand(c)
	}

	layoutName, err = layout.ParseLayout(layoutName)
	if err != nil {
		return err
	}

	plan, err := layout.Calculate(layoutName, paneCount)
	if err != nil {
		return err
	}

	sessionName := fmt.Sprintf("mt-%d", time.Now().Unix())

	paneIDs, err := createSession(sessionName, plan)
	if err != nil {
		return err
	}

	// Label and configure each pane.
	for i, paneID := range paneIDs {
		label := fmt.Sprintf("shell-%d", i+1)
		command := ""

		if i < len(specs) {
			if specs[i].Name != "" {
				label = specs[i].Name
			}
			command = specs[i].Command
		}

		// Set pane title (shows in border).
		_ = tmux.RenamePane(sessionName, paneID, label)

		// Set environment variables for Copilot awareness.
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_SESSION", sessionName)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_ID", paneID)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_NAME", label)

		// Clear the pane after env setup so it looks clean.
		_ = tmux.SendCommand(sessionName, paneID, "clear")

		if command != "" {
			_ = tmux.SendCommand(sessionName, paneID, command)
		}
	}

	// Enable broadcast mode if requested.
	if flagSync {
		_ = tmux.SetSyncPanes(sessionName, true)
	}

	fmt.Printf("✦ multiterm — %d panes [%s] session: %s\n", paneCount, layoutName, sessionName)
	fmt.Println("  Click any pane │ Ctrl-b A: add pane │ Ctrl-b B: sync │ Ctrl-b d: detach")

	_ = tmux.SelectPane(sessionName, paneIDs[0])
	return tmux.AttachSession(sessionName)
}

// createSession creates a tmux session and splits panes according to the plan.
func createSession(sessionName string, plan *layout.LayoutPlan) ([]string, error) {
	if err := tmux.NewSession(sessionName); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	tmux.ConfigureSession(sessionName)

	firstPane, err := tmux.FirstPaneID(sessionName)
	if err != nil {
		_ = tmux.KillSession(sessionName)
		return nil, fmt.Errorf("failed to get first pane: %w", err)
	}

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
			return nil, fmt.Errorf("failed to create pane: %w", err)
		}
		paneIDs = append(paneIDs, newPaneID)
	}

	if plan.TmuxLayoutName != "" {
		if err := tmux.SelectLayout(sessionName, plan.TmuxLayoutName); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not apply layout: %v\n", err)
		}
	}

	return paneIDs, nil
}

// runSSHMultiHost opens one pane per SSH host.
func runSSHMultiHost(cfg *config.Config) error {
	hosts := strings.Split(flagHosts, ",")
	for i, h := range hosts {
		hosts[i] = strings.TrimSpace(h)
	}

	paneCount := len(hosts)
	layoutName := cfg.Defaults.Layout
	if flagLayout != "" {
		layoutName = flagLayout
	}

	layoutName, err := layout.ParseLayout(layoutName)
	if err != nil {
		return err
	}

	plan, err := layout.Calculate(layoutName, paneCount)
	if err != nil {
		return err
	}

	sessionName := fmt.Sprintf("mt-%d", time.Now().Unix())
	paneIDs, err := createSession(sessionName, plan)
	if err != nil {
		return err
	}

	for i, paneID := range paneIDs {
		if i >= len(hosts) {
			break
		}
		host := hosts[i]
		// Extract short hostname for label.
		label := host
		if at := strings.LastIndex(host, "@"); at >= 0 {
			label = host[at+1:]
		}

		_ = tmux.RenamePane(sessionName, paneID, label)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_SESSION", sessionName)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_ID", paneID)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_NAME", label)
		_ = tmux.SendCommand(sessionName, paneID, "clear")
		_ = tmux.SendCommand(sessionName, paneID, fmt.Sprintf("ssh %s", host))
	}

	if flagSync {
		_ = tmux.SetSyncPanes(sessionName, true)
	}

	fmt.Printf("✦ multiterm — %d hosts [%s] session: %s\n", paneCount, layoutName, sessionName)
	fmt.Println("  Click any pane │ Ctrl-b A: add pane │ Ctrl-b B: sync │ Ctrl-b d: detach")

	_ = tmux.SelectPane(sessionName, paneIDs[0])
	return tmux.AttachSession(sessionName)
}

// runCopilotProfile launches the built-in Copilot profile:
// main-side layout with Copilot CLI in the main pane and shells on the side.
func runCopilotProfile(cfg *config.Config) error {
	paneCount := 3
	layoutName := "main-side"

	plan, err := layout.Calculate(layoutName, paneCount)
	if err != nil {
		return err
	}

	sessionName := fmt.Sprintf("mt-%d", time.Now().Unix())
	paneIDs, err := createSession(sessionName, plan)
	if err != nil {
		return err
	}

	labels := []string{"copilot", "shell-1", "shell-2"}
	for i, paneID := range paneIDs {
		label := labels[i]
		_ = tmux.RenamePane(sessionName, paneID, label)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_SESSION", sessionName)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_ID", paneID)
		_ = tmux.SetEnv(sessionName, paneID, "MULTITERM_PANE_NAME", label)
		_ = tmux.SendCommand(sessionName, paneID, "clear")
	}

	// Launch Copilot CLI in the main pane.
	_ = tmux.SendCommand(sessionName, paneIDs[0], "copilot")

	fmt.Printf("✦ multiterm — copilot profile [main-side] session: %s\n", sessionName)
	fmt.Println("  Copilot in main pane │ Click any pane │ Ctrl-b d: detach")

	_ = tmux.SelectPane(sessionName, paneIDs[0])
	return tmux.AttachSession(sessionName)
}
