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

package cluster

import (
	"go.bytebuilders.dev/ace/pkg/config"
	"go.bytebuilders.dev/ace/pkg/printer"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"

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

var defaultFeatureSet = []clustermodel.FeatureSet{
	{
		Name:     "opscenter-core",
		Features: []string{"kube-ui-server"},
	},
}
