package config

import (
	"github.com/spf13/cobra"
)

func NewCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "config",
		Short:             "Manage configuration for the CLI",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(newCmdView())
	cmd.AddCommand(newCmdSet())

	return cmd
}
