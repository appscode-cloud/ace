package cluster

import (
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
)

func newCmdList(f *config.Factory) *cobra.Command {
	listOptions := ace.ClusterListOptions{}
	cmd := &cobra.Command{
		Use:               "list",
		Short:             "List cluster managed by ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := listClusters(f, &listOptions)
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
	cmd.Flags().StringVar(&listOptions.Provider, "provider", "", "List cluster only for this provider")

	return cmd
}

func listClusters(f *config.Factory, opts *ace.ClusterListOptions) (*v1alpha1.ClusterInfoList, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	clusters, err := c.ListClusters(opts)
	if err != nil {
		return nil, err
	}
	for i := range clusters.Items {
		cluster, err := c.GetCluster(ace.ClusterGetOptions{
			Name: clusters.Items[i].Spec.Name,
		})
		if err != nil {
			return nil, err
		}
		clusters.Items[i].Status = cluster.Status
	}
	return clusters, nil
}
