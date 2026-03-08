package cmd

import (
	"fmt"

	"github.com/gilbertappiah/multiterm/internal/config"
	"github.com/spf13/cobra"
)

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default ~/.multiterm.yaml config file",
	Long:  "Generate a starter configuration file with example profiles at ~/.multiterm.yaml.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}
		fmt.Printf("✓ Created %s\n", config.ConfigPath())
		return nil
	},
}
