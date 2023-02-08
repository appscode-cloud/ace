package cluster

import (
	"errors"
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
)

func newCmdConnect(f *config.Factory) *cobra.Command {
	var clusterName, credential string
	cmd := &cobra.Command{
		Use:               "connect",
		Short:             "Connect with a cluster imported by peers",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := connectCluster(f, clusterName, credential)
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
	cmd.Flags().StringVar(&clusterName, "name", "", "Name of the cluster to get")
	cmd.Flags().StringVar(&credential, "credential", "", "Name of the credential to use to connect with the cluster")
	return cmd
}

func connectCluster(f *config.Factory, name, credential string) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.ConnectCluster(ace.ClusterConnectOptions{
		Name:       name,
		Credential: credential,
	})
}
