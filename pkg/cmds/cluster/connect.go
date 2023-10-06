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
	"errors"
	"fmt"
	"os"

	"go.bytebuilders.dev/cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
)

func newCmdConnect(f *config.Factory) *cobra.Command {
	opts := clustermodel.ConnectOptions{}
	var kubeConfigPath string
	cmd := &cobra.Command{
		Use:               "connect",
		Short:             "Connect with a cluster imported by peers",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if kubeConfigPath != "" {
				data, err := os.ReadFile(kubeConfigPath)
				if err != nil {
					return fmt.Errorf("failed to read Kubeconfig file. Reason: %w", err)
				}
				opts.KubeConfig = string(data)
			}
			_, err := connectCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Provided cluster does not exist. Please provide a valid cluster name.")
					return nil
				}
				return fmt.Errorf("failed to connect with cluster. Reason: %w", err)
			}
			fmt.Println("Successfully connected to the cluster")
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().StringVar(&opts.Credential, "credential", "", "Name of the credential to use to connect with the cluster")
	cmd.Flags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path of the kubeconfig file")
	return cmd
}

func connectCluster(f *config.Factory, opts clustermodel.ConnectOptions) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.ConnectCluster(opts)
}
