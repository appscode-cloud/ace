package cluster

import (
	"fmt"

	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func NewCmdListClusters(f *config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list",
		Short:             "List cluster managed by ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := listClusters(f)
			if err != nil {
				return fmt.Errorf("failed to list clusters. Reason: %w", err)
			}
			if len(clusters.Items) == 0 {
				fmt.Println("No cluster found.")
				return nil
			}
			return printClusterList(clusters.Items)
		},
	}
	return cmd
}

func listClusters(f *config.Factory) (*v1alpha1.ClusterInfoList, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.ListClusters()
}
