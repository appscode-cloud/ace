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
	"errors"
	"fmt"

	"go.bytebuilders.dev/ace/pkg/config"

	"github.com/spf13/cobra"
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
