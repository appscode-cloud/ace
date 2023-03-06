package cluster

import (
	"errors"
	"fmt"
	"sync"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	ace "go.bytebuilders.dev/client"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func newCmdReconfigure(f *config.Factory) *cobra.Command {
	opts := clustermodel.ReconfigureOptions{}
	cmd := &cobra.Command{
		Use:               "reconfigure",
		Short:             "Re-install cluster components to fix common issues",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Components.FeatureSets = defaultFeatureSet
			err := reconfigureCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Provided cluster does not exist. Please provide a valid cluster name.")
					return nil
				}
				return fmt.Errorf("failed to reconfigure cluster. Reason: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "install-license-server", true, "Specify whether to install license-server or not (default true).")
	return cmd
}

func reconfigureCluster(f *config.Factory, opts clustermodel.ReconfigureOptions) error {
	fmt.Println("Reconfiguring cluster......")
	c, err := f.Client()
	if err != nil {
		return err
	}
	nc, err := c.NewNatsConnection("ace-cli")
	if err != nil {
		return err
	}
	defer nc.Close()

	responseID := xid.New().String()
	wg := sync.WaitGroup{}
	wg.Add(1)
	done := f.Canceller()
	go func() {
		err := printer.PrintNATSJobSteps(&wg, nc, responseID, done)
		if err != nil {
			fmt.Println("Failed to log reconfigure steps. Reason: ", err)
		}
	}()

	_, err = c.ReconfigureCluster(opts, responseID)
	if err != nil {
		close(done)
		return err
	}
	wg.Wait()

	return nil
}
