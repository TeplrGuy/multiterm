package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const SessionPrefix = "mt-"

// run executes a tmux command and returns its combined output.
func run(args ...string) (string, error) {
	tmux, err := TmuxPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(tmux, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// runSilent executes a tmux command and discards the output.
func runSilent(args ...string) error {
	_, err := run(args...)
	return err
}

// TmuxPath returns the absolute path to the tmux binary.
func TmuxPath() (string, error) {
	p, err := exec.LookPath("tmux")
	if err != nil {
		return "", fmt.Errorf("tmux not found in PATH: %w", err)
	}
	return p, nil
}

// IsInstalled reports whether tmux is available on the system.
func IsInstalled() bool {
	_, err := TmuxPath()
	return err == nil
}

// NewSession creates a new detached tmux session with the given name.
func NewSession(name string) error {
	return runSilent("new-session", "-d", "-s", name)
}

// NewSessionWithEnv creates a new detached tmux session with extra environment variables.
func NewSessionWithEnv(name string, env map[string]string) error {
	args := []string{"new-session", "-d", "-s", name}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	return runSilent(args...)
}

// KillSession destroys the tmux session with the given name.
func KillSession(name string) error {
	return runSilent("kill-session", "-t", name)
}

// SessionExists reports whether a tmux session with the given name exists.
func SessionExists(name string) bool {
	err := runSilent("has-session", "-t", name)
	return err == nil
}

// ListSessions returns the names of all tmux sessions prefixed with "mt-".
func ListSessions() ([]string, error) {
	out, err := run("list-sessions", "-F", "#{session_name}")
	if err != nil {
		if strings.Contains(err.Error(), "no server running") ||
			strings.Contains(err.Error(), "no sessions") ||
			strings.Contains(err.Error(), "error connecting") {
			return nil, nil
		}
		return nil, err
	}

	var sessions []string
	for _, line := range strings.Split(out, "\n") {
		s := strings.TrimSpace(line)
		if s != "" && strings.HasPrefix(s, SessionPrefix) {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}

// AttachSession attaches the current terminal to the named tmux session.
// It replaces the current process via syscall.Exec so the user is placed
// directly inside tmux.
func AttachSession(name string) error {
	tmux, err := TmuxPath()
	if err != nil {
		return err
	}

	argv := []string{"tmux", "attach-session", "-t", name}
	return syscall.Exec(tmux, argv, os.Environ())
}

// SplitVertical splits the target pane vertically (side-by-side) and returns
// the ID of the newly created pane (e.g. "%5").
func SplitVertical(session string, paneID string) (string, error) {
	target := paneTarget(session, paneID)
	out, err := run("split-window", "-h", "-t", target, "-P", "-F", "#{pane_id}")
	if err != nil {
		return "", fmt.Errorf("split vertical in %s: %w", target, err)
	}
	return strings.TrimSpace(out), nil
}

// SplitHorizontal splits the target pane horizontally (stacked) and returns
// the ID of the newly created pane (e.g. "%5").
func SplitHorizontal(session string, paneID string) (string, error) {
	target := paneTarget(session, paneID)
	out, err := run("split-window", "-v", "-t", target, "-P", "-F", "#{pane_id}")
	if err != nil {
		return "", fmt.Errorf("split horizontal in %s: %w", target, err)
	}
	return strings.TrimSpace(out), nil
}

// SendCommand sends a command string to the specified pane and presses Enter.
func SendCommand(session string, paneID string, command string) error {
	target := paneTarget(session, paneID)
	return runSilent("send-keys", "-t", target, command, "Enter")
}

// SetEnv sets an environment variable in a specific pane.
func SetEnv(session string, paneID string, key, value string) error {
	target := paneTarget(session, paneID)
	envCmd := fmt.Sprintf("export %s=%q", key, value)
	return runSilent("send-keys", "-t", target, envCmd, "Enter")
}

// RenamePane sets a custom title on a pane using the pane_title escape sequence.
func RenamePane(session string, paneID string, title string) error {
	target := paneTarget(session, paneID)
	// Use OSC escape sequence to set pane title.
	escape := fmt.Sprintf("printf '\\033]2;%s\\033\\\\'", title)
	return runSilent("send-keys", "-t", target, escape, "Enter")
}

// SetSyncPanes toggles synchronized input across all panes in the window.
func SetSyncPanes(session string, on bool) error {
	value := "off"
	if on {
		value = "on"
	}
	target := fmt.Sprintf("%s:0", session)
	return runSilent("set-window-option", "-t", target, "synchronize-panes", value)
}

// BindKey binds a key in the session.
func BindKey(session string, key string, tmuxCmd string) error {
	return runSilent("bind-key", "-T", "prefix", key, tmuxCmd)
}

// FirstPaneID returns the pane ID of the first pane in the session.
func FirstPaneID(session string) (string, error) {
	out, err := run("list-panes", "-t", session+":0", "-F", "#{pane_id}")
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no panes found in session %s", session)
	}
	return strings.TrimSpace(lines[0]), nil
}

// paneTarget builds the correct tmux target string. Pane IDs starting with
// "%" are global and can be used directly; otherwise use session:window.pane.
func paneTarget(session string, paneID string) string {
	if strings.HasPrefix(paneID, "%") {
		return paneID
	}
	return fmt.Sprintf("%s:0.%s", session, paneID)
}

// SelectLayout applies a tmux layout to the first window of the session.
func SelectLayout(session string, layout string) error {
	target := fmt.Sprintf("%s:0", session)
	return runSilent("select-layout", "-t", target, layout)
}

// SetOption sets a tmux session option.
func SetOption(session string, option string, value string) error {
	return runSilent("set-option", "-t", session, option, value)
}

// SetGlobalOption sets a tmux server-wide option.
func SetGlobalOption(option string, value string) error {
	return runSilent("set-option", "-g", option, value)
}

// ConfigureSession applies sensible defaults so every pane is independently
// interactive: mouse mode on, pane borders visible, and status bar info.
func ConfigureSession(session string) {
	_ = SetOption(session, "mouse", "on")
	_ = SetOption(session, "pane-border-style", "fg=colour240")
	_ = SetOption(session, "pane-active-border-style", "fg=colour51,bold")
	_ = SetOption(session, "pane-border-status", "top")
	_ = SetOption(session, "pane-border-format", " #{pane_title} ")
	_ = SetOption(session, "status-left", " ✦ multiterm ")
	_ = SetOption(session, "status-style", "bg=colour236,fg=colour75")
	_ = SetOption(session, "status-right", " ^b A: add pane │ ^b B: sync │ ^b d: detach ")
	_ = SetOption(session, "status-right-length", "60")

	// Bind Ctrl-b B to toggle synchronize-panes.
	_ = runSilent("bind-key", "-T", "prefix", "B",
		"set-window-option", "synchronize-panes")

	// Bind Ctrl-b A to split a new pane and auto-tile.
	_ = runSilent("bind-key", "-T", "prefix", "A",
		"split-window", "-v", "\\;",
		"select-layout", "tiled")
}

// SelectPane sets the active pane in the session.
func SelectPane(session string, paneID string) error {
	target := paneTarget(session, paneID)
	return runSilent("select-pane", "-t", target)
}

// CapturePane captures the visible content of a pane.
// Returns up to `lines` lines from the bottom of the pane scrollback.
func CapturePane(session string, paneID string, lines int) (string, error) {
	target := paneTarget(session, paneID)
	start := fmt.Sprintf("-%d", lines)
	out, err := run("capture-pane", "-p", "-t", target, "-S", start)
	if err != nil {
		return "", fmt.Errorf("capture pane %s: %w", target, err)
	}
	return out, nil
}

// ClosePane closes a specific pane in the session.
func ClosePane(session string, paneID string) error {
	target := paneTarget(session, paneID)
	return runSilent("send-keys", "-t", target, "exit", "Enter")
}

// ListPaneInfo returns info about all panes in a session.
type PaneInfo struct {
	ID    string
	Index int
	Title string
	Cmd   string
}

func ListPanes(session string) ([]PaneInfo, error) {
	out, err := run("list-panes", "-t", session+":0", "-F",
		"#{pane_id}\t#{pane_index}\t#{pane_title}\t#{pane_current_command}")
	if err != nil {
		return nil, err
	}

	var panes []PaneInfo
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "\t", 4)
		if len(parts) < 4 {
			continue
		}
		idx := 0
		fmt.Sscanf(parts[1], "%d", &idx)
		panes = append(panes, PaneInfo{
			ID:    parts[0],
			Index: idx,
			Title: parts[2],
			Cmd:   parts[3],
		})
	}
	return panes, nil
}
