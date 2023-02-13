package cluster

import (
	"fmt"
	"sync"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	ace "go.bytebuilders.dev/client"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func newCmdImport(f *config.Factory) *cobra.Command {
	opts := ace.ClusterImportOptions{}
	cmd := &cobra.Command{
		Use:               "import",
		Short:             "Import a cluster to ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := importCluster(f, opts)
			if err != nil {
				return fmt.Errorf("failed to import cluster. Reason: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Provider.Provider, "provider", "", "Name of the cluster provider")
	cmd.Flags().StringVar(&opts.Provider.Credential, "credential", "", "Name of the credential with access to the provider APIs")
	cmd.Flags().StringVar(&opts.Provider.ClusterID, "id", "", "Provider specific cluster ID")

	cmd.Flags().StringVar(&opts.BasicInfo.DisplayName, "display-name", "", "Display name of the cluster")
	cmd.Flags().StringVar(&opts.BasicInfo.Name, "name", "", "Unique name across all imported clusters of all provider")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "install-license-server", true, "Specify whether to install license-server or not (default true).")

	return cmd
}

func importCluster(f *config.Factory, opts ace.ClusterImportOptions) error {
	fmt.Println("Importing cluster......")
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
			fmt.Println("Failed to log the import steps. Reason: ", err)
		}
	}()

	opts.ResponseID = responseID
	_, err = c.ImportCluster(opts)
	if err != nil {
		close(done)
		return err
	}
	wg.Wait()

	return nil
}
