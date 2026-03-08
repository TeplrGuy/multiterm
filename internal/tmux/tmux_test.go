package tmux

import (
	"strings"
	"testing"
)

func TestPaneTarget(t *testing.T) {
	tests := []struct {
		name     string
		session  string
		paneID   string
		expected string
	}{
		{"global pane ID", "mt-123", "%5", "%5"},
		{"relative pane ID", "mt-123", "2", "mt-123:0.2"},
		{"empty pane ID", "mt-123", "", "mt-123:0."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paneTarget(tt.session, tt.paneID)
			if result != tt.expected {
				t.Errorf("paneTarget(%q, %q) = %q, want %q", tt.session, tt.paneID, result, tt.expected)
			}
		})
	}
}

func TestSessionPrefix(t *testing.T) {
	if SessionPrefix != "mt-" {
		t.Errorf("SessionPrefix = %q, want %q", SessionPrefix, "mt-")
	}
}

func TestIsInstalled(t *testing.T) {
	// tmux should be installed in test environments
	if !IsInstalled() {
		t.Skip("tmux not installed, skipping")
	}
}

func TestTmuxPath(t *testing.T) {
	path, err := TmuxPath()
	if err != nil {
		t.Skip("tmux not installed, skipping")
	}
	if path == "" {
		t.Error("TmuxPath returned empty string")
	}
}

func TestListSessions_NoServer(t *testing.T) {
	// When no tmux server is running, ListSessions should return nil, nil
	// (graceful handling of "no server running" error).
	// This test may pass or return sessions depending on environment.
	sessions, err := ListSessions()
	if err != nil {
		t.Errorf("ListSessions returned unexpected error: %v", err)
	}
	// sessions may be nil or contain entries — both are valid
	_ = sessions
}

func TestSessionExists_Nonexistent(t *testing.T) {
	if SessionExists("nonexistent-session-abc123") {
		t.Error("SessionExists returned true for nonexistent session")
	}
}

// Integration tests that create real tmux sessions.
// These test the full lifecycle: create → list → capture → close.
func TestSessionLifecycle(t *testing.T) {
	if !IsInstalled() {
		t.Skip("tmux not installed")
	}

	session := "mt-test-lifecycle"

	// Cleanup in case of prior failed test.
	_ = KillSession(session)

	// Create session.
	if err := NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer KillSession(session)

	// Verify exists.
	if !SessionExists(session) {
		t.Fatal("session should exist after creation")
	}

	// Get first pane.
	firstPane, err := FirstPaneID(session)
	if err != nil {
		t.Fatalf("FirstPaneID failed: %v", err)
	}
	if !strings.HasPrefix(firstPane, "%") {
		t.Errorf("FirstPaneID returned %q, expected %%N format", firstPane)
	}

	// Split pane.
	newPane, err := SplitHorizontal(session, firstPane)
	if err != nil {
		t.Fatalf("SplitHorizontal failed: %v", err)
	}
	if !strings.HasPrefix(newPane, "%") {
		t.Errorf("SplitHorizontal returned %q, expected %%N format", newPane)
	}

	// List panes.
	panes, err := ListPanes(session)
	if err != nil {
		t.Fatalf("ListPanes failed: %v", err)
	}
	if len(panes) != 2 {
		t.Errorf("expected 2 panes, got %d", len(panes))
	}

	// Send command.
	if err := SendCommand(session, firstPane, "echo multiterm-test-marker"); err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}

	// Capture pane output.
	output, err := CapturePane(session, firstPane, 10)
	if err != nil {
		t.Fatalf("CapturePane failed: %v", err)
	}
	// Output should contain something (may or may not have our marker yet).
	_ = output

	// Select layout.
	if err := SelectLayout(session, "tiled"); err != nil {
		t.Errorf("SelectLayout failed: %v", err)
	}

	// Configure session (should not error).
	ConfigureSession(session)

	// Rename pane.
	if err := RenamePane(session, firstPane, "test-pane"); err != nil {
		t.Errorf("RenamePane failed: %v", err)
	}

	// Set env var.
	if err := SetEnv(session, firstPane, "TEST_VAR", "hello"); err != nil {
		t.Errorf("SetEnv failed: %v", err)
	}

	// Select pane.
	if err := SelectPane(session, firstPane); err != nil {
		t.Errorf("SelectPane failed: %v", err)
	}

	// Close pane.
	if err := ClosePane(session, newPane); err != nil {
		t.Errorf("ClosePane failed: %v", err)
	}

	// List sessions should include our session.
	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s == session {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("session %q not found in ListSessions output: %v", session, sessions)
	}

	// Kill session.
	if err := KillSession(session); err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}

	// Verify gone.
	if SessionExists(session) {
		t.Error("session should not exist after kill")
	}
}

func TestSplitVertical(t *testing.T) {
	if !IsInstalled() {
		t.Skip("tmux not installed")
	}

	session := "mt-test-vsplit"
	_ = KillSession(session)

	if err := NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer KillSession(session)

	firstPane, err := FirstPaneID(session)
	if err != nil {
		t.Fatalf("FirstPaneID failed: %v", err)
	}

	newPane, err := SplitVertical(session, firstPane)
	if err != nil {
		t.Fatalf("SplitVertical failed: %v", err)
	}
	if !strings.HasPrefix(newPane, "%") {
		t.Errorf("expected %%N format, got %q", newPane)
	}

	panes, err := ListPanes(session)
	if err != nil {
		t.Fatalf("ListPanes failed: %v", err)
	}
	if len(panes) != 2 {
		t.Errorf("expected 2 panes after vertical split, got %d", len(panes))
	}
}

func TestSyncPanes(t *testing.T) {
	if !IsInstalled() {
		t.Skip("tmux not installed")
	}

	session := "mt-test-sync"
	_ = KillSession(session)

	if err := NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer KillSession(session)

	// Should not error.
	if err := SetSyncPanes(session, true); err != nil {
		t.Errorf("SetSyncPanes(true) failed: %v", err)
	}
	if err := SetSyncPanes(session, false); err != nil {
		t.Errorf("SetSyncPanes(false) failed: %v", err)
	}
}
