package installer

import (
	"github.com/spf13/cobra"
)

func NewCmdInstaller() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Ace Installer Commands",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(newCmdOpenShift())
	cmd.AddCommand(newCmdValidate())

	return cmd
}
