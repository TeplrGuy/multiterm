package cmd

import (
	"fmt"

	mcpserver "github.com/TeplrGuy/multiterm/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start multiterm MCP server (for Copilot CLI integration)",
	Long: `Start a Model Context Protocol (MCP) server on stdio.

This lets GitHub Copilot CLI manage terminal panes through multiterm.
Copilot can create panes, run commands, read output, and more — making
all agent work visible in your terminal.

The server is automatically configured when you use 'multiterm copilot'.
For manual setup, add to your MCP config:

  {
    "multiterm": {
      "type": "stdio",
      "command": "multiterm",
      "args": ["serve"]
    }
  }`,
	SilenceUsage: true,
	RunE:         runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.ErrOrStderr(), "✦ multiterm MCP server starting...")
	return mcpserver.Serve(version)
}
