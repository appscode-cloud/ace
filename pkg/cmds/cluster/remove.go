package cluster

import (
	"errors"
	"fmt"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
)

func newCmdRemove(f *config.Factory) *cobra.Command {
	opts := ace.ClusterRemovalOptions{}
	cmd := &cobra.Command{
		Use:               "remove",
		Short:             "Remove a cluster from ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := removeCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Cluster has been removed already.")
					return nil
				}
				return fmt.Errorf("failed to remove cluster. Reason: %w", err)
			}
			fmt.Println("Successfully removed the cluster.")
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "remove-fluxcd", true, "Specify whether to remove FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "remove-license-server", true, "Specify whether to remove license server or not (default true).")
	return cmd
}

func removeCluster(f *config.Factory, opts ace.ClusterRemovalOptions) error {
	fmt.Println("Removing cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	err = c.RemoveCluster(opts)
	if err != nil {
		return err
	}

	return waitForClusterToBeRemoved(c, opts.Name)
}
