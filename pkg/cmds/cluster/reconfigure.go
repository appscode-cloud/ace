package cluster

import (
	"errors"
	"fmt"

	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func newCmdReconfigure(f *config.Factory) *cobra.Command {
	var clusterName string
	var installFluxCD bool
	cmd := &cobra.Command{
		Use:               "reconfigure",
		Short:             "Re-install cluster components to fix common issues",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := reconfigureCluster(f, clusterName, installFluxCD)
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
	cmd.Flags().StringVar(&clusterName, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&installFluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	return cmd
}

func reconfigureCluster(f *config.Factory, name string, installFluxCD bool) error {
	fmt.Println("Reconfiguring cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	cluster, err := c.ReconfigureCluster(name, installFluxCD)
	if err != nil {
		return err
	}
	return waitForClusterToBeReady(c, cluster.Spec.Name)
}
