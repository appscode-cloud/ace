package config

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func newCmdDelete() *cobra.Command {
	var context string
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete a context for the cli configuration file",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.DeleteContext(context)
			if err != nil {
				if errors.Is(err, config.ErrContextNotFound) {
					fmt.Println("Context does not exist. Nothing to do.")
					return nil
				}
				return err
			}
			fmt.Println("Successfully remove the context.")
			return nil
		},
	}
	cmd.Flags().StringVar(&context, "context", "", "Name of the context to use")
	return cmd
}
