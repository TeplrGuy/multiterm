package mcp

import (
	"context"
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
