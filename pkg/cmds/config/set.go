package config

import (
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"

	"github.com/spf13/cobra"
)

func newCmdSet() *cobra.Command {
	ctx := config.Context{}
	cmd := &cobra.Command{
		Use:               "set",
		Short:             "Create a new context or update existing context in CLI configuration",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.SetContext(ctx)
			if err != nil {
				return err
			}
			fmt.Println("Successfully set the context")
			return nil
		},
	}

	cmd.Flags().StringVar(&ctx.Name, "name", "", "Name of the context")
	cmd.Flags().StringVar(&ctx.Endpoint, "endpoint", "", "API endpoint of this context")
	cmd.Flags().StringVar(&ctx.Token, "token", "", "Token for this endpoint")

	return cmd
}
