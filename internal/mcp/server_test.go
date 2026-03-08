package mcp

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer(t *testing.T) {
	s := NewServer("test-1.0.0")
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		key      string
		expected string
	}{
		{"existing key", map[string]any{"name": "hello"}, "name", "hello"},
		{"missing key", map[string]any{"name": "hello"}, "other", ""},
		{"non-string value", map[string]any{"count": 42}, "count", ""},
		{"nil map", nil, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.args, tt.key)
			if result != tt.expected {
				t.Errorf("getString(%v, %q) = %q, want %q", tt.args, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		key        string
		defaultVal int
		expected   int
	}{
		{"float64 value", map[string]any{"lines": float64(100)}, "lines", 50, 100},
		{"int value", map[string]any{"lines": 75}, "lines", 50, 75},
		{"missing key uses default", map[string]any{}, "lines", 50, 50},
		{"nil map uses default", nil, "lines", 50, 50},
		{"string value uses default", map[string]any{"lines": "abc"}, "lines", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("getInt(%v, %q, %d) = %d, want %d", tt.args, tt.key, tt.defaultVal, result, tt.expected)
			}
		})
	}
}

func TestResolveSession_NoSessions(t *testing.T) {
	_, err := resolveSession("nonexistent-session-xyz")
	if err == nil {
		t.Error("expected error for nonexistent session, got nil")
	}
}

func TestHandleListPanes_NoSession(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"session": "nonexistent-mt-session",
	}

	result, err := handleListPanes(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for nonexistent session")
	}
}

func TestHandleCreatePane_NoSession(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"session": "nonexistent-mt-session",
		"name":    "test-pane",
	}

	result, err := handleCreatePane(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for nonexistent session")
	}
}

func TestHandleRunInPane_MissingArgs(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handleRunInPane(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing args")
	}
}

func TestHandleReadPane_MissingPaneID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handleReadPane(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing pane_id")
	}
}

func TestHandleClosePane_MissingPaneID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handleClosePane(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing pane_id")
	}
}

func TestHandleBroadcast_MissingCommand(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handleBroadcast(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing command")
	}
}

func TestAutoSessionName(t *testing.T) {
	if autoSessionName != "mt-copilot" {
		t.Errorf("autoSessionName = %q, want %q", autoSessionName, "mt-copilot")
	}
}

func TestResolveSession_ExplicitNotFound(t *testing.T) {
	_, err := resolveSession("mt-does-not-exist-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent explicit session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestResolveSession_AutoCreate(t *testing.T) {
	// Skip if tmux is not available
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available")
	}

	// Clean up any existing auto-session
	exec.Command("tmux", "kill-session", "-t", autoSessionName).Run()

	// Resolve with empty string should auto-create
	session, err := resolveSession("")
	if err != nil {
		// May fail if other mt-* sessions exist; that's OK
		if strings.Contains(err.Error(), "auto-create") {
			t.Fatalf("auto-create failed: %v", err)
		}
		// If it found an existing session, that's fine too
		return
	}

	if session == "" {
		t.Fatal("resolveSession returned empty session name")
	}

	// If it auto-created, clean up
	if session == autoSessionName {
		exec.Command("tmux", "kill-session", "-t", autoSessionName).Run()
	}
}
