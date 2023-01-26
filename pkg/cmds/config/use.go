package config

import (
	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func newCmdUse() *cobra.Command {
	var context string
	cmd := &cobra.Command{
		Use:               "use",
		Short:             "Use provided context as current context",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.SetCurrentContext(context)
		},
	}
	cmd.Flags().StringVar(&context, "context", "", "Name of the context to use")
	return cmd
}
