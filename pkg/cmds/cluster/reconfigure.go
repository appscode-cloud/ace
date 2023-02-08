package cluster

import (
	"errors"
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
)

func newCmdReconfigure(f *config.Factory) *cobra.Command {
	opts := ace.ClusterReconfigureOptions{}
	cmd := &cobra.Command{
		Use:               "reconfigure",
		Short:             "Re-install cluster components to fix common issues",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := reconfigureCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Provided cluster does not exist. Please provide a valid cluster name.")
					return nil
				}
				return fmt.Errorf("failed to reconfigure cluster. Reason: %w", err)
			}
			fmt.Println("Successfully reconfigured the cluster.")
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "install-license-server", true, "Specify whether to install license-server or not (default true).")
	return cmd
}

func reconfigureCluster(f *config.Factory, opts ace.ClusterReconfigureOptions) error {
	fmt.Println("Reconfiguring cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	cluster, err := c.ReconfigureCluster(opts)
	if err != nil {
		return err
	}
	return waitForClusterToBeReady(c, cluster.Spec.Name)
}
