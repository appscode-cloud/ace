/*
Copyright AppsCode Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"

	"github.com/spf13/cobra"
)

func NewCmdCluster(f *config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cluster",
		Short:             "Manage clusters in ACE",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdCheck(f))
	cmd.AddCommand(newCmdImport(f))
	cmd.AddCommand(newCmdGet(f))
	cmd.AddCommand(newCmdConnect(f))
	cmd.AddCommand(newCmdReconfigure(f))
	cmd.AddCommand(newCmdRemove(f))

	cmd.PersistentFlags().StringVarP(&printer.OutputFormat, "output", "o", "", "Output format (any of json,yaml,table). Default is table.")
	return cmd
}
