package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
)

type importOptions struct {
	provider      ace.ProviderOptions
	basicInfo     ace.ClusterBasicInfo
	installFluxCD bool
}

func newCmdImport(f *config.Factory) *cobra.Command {
	opts := importOptions{}
	cmd := &cobra.Command{
		Use:               "import",
		Short:             "Import a cluster to ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := importCluster(f, opts)
			if err != nil {
				return fmt.Errorf("failed to import cluster. Reason: %w", err)
			}
			fmt.Println("Successfully imported the cluster.")
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.provider.Provider, "provider", "", "Name of the cluster provider")
	cmd.Flags().StringVar(&opts.provider.Credential, "credential", "", "Name of the credential with access to the provider APIs")
	cmd.Flags().StringVar(&opts.provider.ClusterID, "id", "", "Provider specific cluster ID")

	cmd.Flags().StringVar(&opts.basicInfo.DisplayName, "display-name", "", "Display name of the cluster")
	cmd.Flags().StringVar(&opts.basicInfo.Name, "name", "", "Unique name across all imported clusters of all provider")
	cmd.Flags().BoolVar(&opts.installFluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")

	return cmd
}

func importCluster(f *config.Factory, opts importOptions) error {
	c, err := f.Client()
	if err != nil {
		return err
	}
	fmt.Println("Importing cluster......")
	cluster, err := c.ImportCluster(opts.basicInfo, opts.provider, opts.installFluxCD)
	if err != nil {
		return err
	}
	return waitForClusterToBeReady(c, cluster.Spec.Name)
}
