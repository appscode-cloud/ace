package cluster

import (
	"errors"
	"fmt"

	ace "go.bytebuilders.dev/client"

	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func newCmdGet(f *config.Factory) *cobra.Command {
	var clusterName string
	cmd := &cobra.Command{
		Use:               "get",
		Short:             "Get a particular cluster information",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := getCluster(f, clusterName)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Cluster does not exist.")
					return nil
				}
				return fmt.Errorf("failed to get the cluster information. Reason: %w", err)
			}
			return printCluster(cluster)
		},
	}
	cmd.Flags().StringVar(&clusterName, "name", "", "Name of the cluster to get")
	return cmd
}

func getCluster(f *config.Factory, name string) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.GetCluster(name)
}
