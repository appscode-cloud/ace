package installer

import (
	"github.com/spf13/cobra"
)

func newCmdOpenShift() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "openshift",
		Short:             "Configure OpenShift scc",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
