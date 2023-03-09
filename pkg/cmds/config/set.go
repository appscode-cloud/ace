/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
