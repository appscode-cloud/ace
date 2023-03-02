package cluster

import (
	"fmt"
	"os"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
)

func newCmdCheck(f *config.Factory) *cobra.Command {
	opts := clustermodel.CheckOptions{}
	var kubeConfigPath string
	cmd := &cobra.Command{
		Use:               "check",
		Short:             "Check whether a cluster has been imported already or not",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if kubeConfigPath != "" {
				data, err := os.ReadFile(kubeConfigPath)
				if err != nil {
					return fmt.Errorf("failed to read Kubeconfig file. Reason: %w", err)
				}
				opts.Provider.KubeConfig = string(data)
			}
			cluster, err := checkClusterExistence(f, opts)
			if err != nil {
				return fmt.Errorf("failed to check cluster existence. Reason: %w", err)
			}
			if cluster.Status.Phase == v1alpha1.ClusterPhaseNotImported {
				fmt.Println("Cluster hasn't been imported yet.")
				return nil
			}
			return printer.PrintCluster(cluster)
		},
	}
	cmd.Flags().StringVar(&opts.Provider.Name, "provider", "", "Name of the cluster provider")
	cmd.Flags().StringVar(&opts.Provider.Credential, "credential", "", "Name of the credential with access to the provider APIs")
	cmd.Flags().StringVar(&opts.Provider.ClusterID, "id", "", "Provider specific cluster ID")
	cmd.Flags().StringVar(&opts.Provider.Project, "project", "", "Project where the cluster belong (use for GKE)")
	cmd.Flags().StringVar(&opts.Provider.Region, "region", "", "Region or location of the cluster")
	cmd.Flags().StringVar(&opts.Provider.ResourceGroup, "resource-group", "", "Resource group of the cluster (use for AKS)")
	cmd.Flags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path of the kubeconfig file")

	return cmd
}

func checkClusterExistence(f *config.Factory, opts clustermodel.CheckOptions) (*v1alpha1.ClusterInfo, error) {
	c, err := f.Client()
	if err != nil {
		return nil, err
	}
	return c.CheckClusterExistence(opts)
}
