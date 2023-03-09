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
