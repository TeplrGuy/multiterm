package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TeplrGuy/multiterm/internal/tmux"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates a new MCP server with all multiterm tools registered.
func NewServer(version string) *server.MCPServer {
	s := server.NewMCPServer(
		"multiterm",
		version,
		server.WithToolCapabilities(false),
	)

	s.AddTool(createPaneTool(), handleCreatePane)
	s.AddTool(runInPaneTool(), handleRunInPane)
	s.AddTool(readPaneTool(), handleReadPane)
	s.AddTool(listPanesTool(), handleListPanes)
	s.AddTool(closePaneTool(), handleClosePane)
	s.AddTool(broadcastTool(), handleBroadcast)

	return s
}

// Serve starts the MCP server on stdio.
func Serve(version string) error {
	s := NewServer(version)
	return server.ServeStdio(s)
}

// resolveSession finds the target session name.
// If session is provided, use it. Otherwise find the most recent mt-* session.
func resolveSession(session string) (string, error) {
	if session != "" {
		if !tmux.SessionExists(session) {
			return "", fmt.Errorf("session %q not found", session)
		}
		return session, nil
	}

	sessions, err := tmux.ListSessions()
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}
	if len(sessions) == 0 {
		return "", fmt.Errorf("no multiterm sessions running — start one with: multiterm")
	}
	return sessions[len(sessions)-1], nil
}

func getString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}

// --- Tool Definitions ---

func createPaneTool() mcp.Tool {
	return mcp.NewTool("create_pane",
		mcp.WithDescription("Create a new terminal pane in the multiterm session. Use this to make agent work visible — each task gets its own pane."),
		mcp.WithString("name",
			mcp.Description("Label for the pane border (e.g. 'test-runner', 'build', 'logs')"),
		),
		mcp.WithString("command",
			mcp.Description("Command to run in the new pane (e.g. 'npm test', 'go build ./...')"),
		),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

func runInPaneTool() mcp.Tool {
	return mcp.NewTool("run_in_pane",
		mcp.WithDescription("Execute a command in an existing pane. Sends the command and presses Enter."),
		mcp.WithString("pane_id",
			mcp.Required(),
			mcp.Description("Pane ID (e.g. '%5') from list_panes output"),
		),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Command to execute in the pane"),
		),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

func readPaneTool() mcp.Tool {
	return mcp.NewTool("read_pane",
		mcp.WithDescription("Capture visible output from a pane. Use this to see what a command produced — great for checking test results, build output, or error messages."),
		mcp.WithString("pane_id",
			mcp.Required(),
			mcp.Description("Pane ID (e.g. '%5') from list_panes output"),
		),
		mcp.WithNumber("lines",
			mcp.Description("Number of lines to capture from the bottom (default: 50)"),
		),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

func listPanesTool() mcp.Tool {
	return mcp.NewTool("list_panes",
		mcp.WithDescription("List all panes in the multiterm session with their IDs, names, and running commands."),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

func closePaneTool() mcp.Tool {
	return mcp.NewTool("close_pane",
		mcp.WithDescription("Close a specific pane by sending 'exit' to it."),
		mcp.WithString("pane_id",
			mcp.Required(),
			mcp.Description("Pane ID (e.g. '%5') to close"),
		),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

func broadcastTool() mcp.Tool {
	return mcp.NewTool("broadcast",
		mcp.WithDescription("Send the same command to ALL panes simultaneously. Useful for cluster-wide operations."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Command to broadcast to all panes"),
		),
		mcp.WithString("session",
			mcp.Description("Target session name. If omitted, uses the most recent multiterm session."),
		),
	)
}

// --- Tool Handlers ---

func handleCreatePane(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name := getString(args, "name")
	command := getString(args, "command")
	sessionArg := getString(args, "session")

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get current pane list for labeling.
	panes, err := tmux.ListPanes(session)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list panes: %v", err)), nil
	}

	nextIndex := len(panes) + 1
	if name == "" {
		name = fmt.Sprintf("shell-%d", nextIndex)
	}

	// Split from the last pane.
	lastPaneID := panes[len(panes)-1].ID
	newPaneID, err := tmux.SplitHorizontal(session, lastPaneID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to split pane: %v", err)), nil
	}

	// Auto-tile for even distribution.
	_ = tmux.SelectLayout(session, "tiled")

	// Label and configure.
	_ = tmux.RenamePane(session, newPaneID, name)
	_ = tmux.SetEnv(session, newPaneID, "MULTITERM_SESSION", session)
	_ = tmux.SetEnv(session, newPaneID, "MULTITERM_PANE_ID", newPaneID)
	_ = tmux.SetEnv(session, newPaneID, "MULTITERM_PANE_NAME", name)
	_ = tmux.SendCommand(session, newPaneID, "clear")

	if command != "" {
		_ = tmux.SendCommand(session, newPaneID, command)
	}

	result := map[string]string{
		"pane_id": newPaneID,
		"name":    name,
		"session": session,
	}
	b, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(b)), nil
}

func handleRunInPane(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	paneID := getString(args, "pane_id")
	command := getString(args, "command")
	sessionArg := getString(args, "session")

	if paneID == "" || command == "" {
		return mcp.NewToolResultError("pane_id and command are required"), nil
	}

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := tmux.SendCommand(session, paneID, command); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to send command: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Sent command to pane %s: %s", paneID, command)), nil
}

func handleReadPane(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	paneID := getString(args, "pane_id")
	lines := getInt(args, "lines", 50)
	sessionArg := getString(args, "session")

	if paneID == "" {
		return mcp.NewToolResultError("pane_id is required"), nil
	}

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	output, err := tmux.CapturePane(session, paneID, lines)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to capture pane: %v", err)), nil
	}

	return mcp.NewToolResultText(output), nil
}

func handleListPanes(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	sessionArg := getString(args, "session")

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	panes, err := tmux.ListPanes(session)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list panes: %v", err)), nil
	}

	var lines []string
	for _, p := range panes {
		lines = append(lines, fmt.Sprintf("id=%s index=%d name=%q cmd=%s", p.ID, p.Index, p.Title, p.Cmd))
	}

	return mcp.NewToolResultText(strings.Join(lines, "\n")), nil
}

func handleClosePane(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	paneID := getString(args, "pane_id")
	sessionArg := getString(args, "session")

	if paneID == "" {
		return mcp.NewToolResultError("pane_id is required"), nil
	}

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := tmux.ClosePane(session, paneID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to close pane: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Closed pane %s", paneID)), nil
}

func handleBroadcast(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	command := getString(args, "command")
	sessionArg := getString(args, "session")

	if command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	session, err := resolveSession(sessionArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	panes, err := tmux.ListPanes(session)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list panes: %v", err)), nil
	}

	for _, p := range panes {
		_ = tmux.SendCommand(session, p.ID, command)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Broadcast to %d panes: %s", len(panes), command)), nil
}
