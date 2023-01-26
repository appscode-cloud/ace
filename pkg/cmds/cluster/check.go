package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"
)

func newCmdCheck(f *config.Factory) *cobra.Command {
	opts := ace.ProviderOptions{}
	cmd := &cobra.Command{
		Use:               "check",
		Short:             "Check whether a cluster has been imported already or not",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := checkClusterExistence(f, opts)
			if err != nil {
				return fmt.Errorf("failed to check cluster existence. Reason: %w", err)
			}
			if cluster.Status.Phase == v1alpha1.ClusterPhaseNotImported {
				fmt.Println("Cluster hasn't been imported yet.")
				return nil
			}
			return printCluster(cluster)
		},
	}
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Name of the cluster provider")
	cmd.Flags().StringVar(&opts.Credential, "credential", "", "Name of the credential with access to the provider APIs")
	cmd.Flags().StringVar(&opts.ClusterID, "id", "", "Provider specific cluster ID")

	return cmd
}

func checkClusterExistence(f *config.Factory, opts ace.ProviderOptions) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.CheckClusterExistence(opts)
}
