package cluster

import (
	"errors"
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	ace "go.bytebuilders.dev/client"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
)

func newCmdGet(f *config.Factory) *cobra.Command {
	opts := clustermodel.GetOptions{}
	cmd := &cobra.Command{
		Use:               "get",
		Short:             "Get a particular cluster information",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := getCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Cluster does not exist.")
					return nil
				}
				return fmt.Errorf("failed to get the cluster information. Reason: %w", err)
			}
			return printer.PrintCluster(cluster)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	return cmd
}

func getCluster(f *config.Factory, opts clustermodel.GetOptions) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.GetCluster(opts)
}
