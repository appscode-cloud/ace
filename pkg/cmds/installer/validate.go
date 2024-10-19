package installer

import (
	"github.com/spf13/cobra"
)

func newCmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "validate",
		Short:             "Validate installer options",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
