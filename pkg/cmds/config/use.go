package config

import (
	"go.bytebuilders.dev/ace-cli/pkg/config"

	"github.com/spf13/cobra"
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
