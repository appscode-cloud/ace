package cluster

import (
	"errors"
	"fmt"

	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
)

func newCmdRemove(f *config.Factory) *cobra.Command {
	var clusterName string
	var removeFluxCD bool
	cmd := &cobra.Command{
		Use:               "remove",
		Short:             "Remove a cluster from ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := removeCluster(f, clusterName, removeFluxCD)
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
	cmd.Flags().StringVar(&clusterName, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&removeFluxCD, "remove-fluxcd", true, "Specify whether to remove FluxCD or not (default true).")
	return cmd
}

func removeCluster(f *config.Factory, name string, removeFluxCD bool) error {
	fmt.Println("Removing cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	err = c.RemoveCluster(name, removeFluxCD)
	if err != nil {
		return err
	}

	return waitForClusterToBeRemoved(c, name)
}
