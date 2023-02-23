package cluster

import (
	"fmt"
	"os"
	"sync"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	"go.bytebuilders.dev/ace-cli/pkg/printer"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func newCmdImport(f *config.Factory) *cobra.Command {
	opts := clustermodel.ImportOptions{}
	var kubeConfigPath string
	cmd := &cobra.Command{
		Use:               "import",
		Short:             "Import a cluster to ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if kubeConfigPath != "" {
				data, err := os.ReadFile(kubeConfigPath)
				if err != nil {
					return fmt.Errorf("failed to read Kubeconfig file. Reason: %w", err)
				}
				opts.Provider.KubeConfig = string(data)
			}
			err := importCluster(f, opts)
			if err != nil {
				return fmt.Errorf("failed to import cluster. Reason: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Provider.Name, "provider", "", "Name of the cluster provider")
	cmd.Flags().StringVar(&opts.Provider.Credential, "credential", "", "Name of the credential with access to the provider APIs")
	cmd.Flags().StringVar(&opts.Provider.ClusterID, "id", "", "Provider specific cluster ID")
	cmd.Flags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path of the kubeconfig file")

	cmd.Flags().StringVar(&opts.BasicInfo.DisplayName, "display-name", "", "Display name of the cluster")
	cmd.Flags().StringVar(&opts.BasicInfo.Name, "name", "", "Unique name across all imported clusters of all provider")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.LicenseServer, "install-license-server", true, "Specify whether to install license-server or not (default true).")

	return cmd
}

func importCluster(f *config.Factory, opts clustermodel.ImportOptions) error {
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

	_, err = c.ImportCluster(opts, responseID)
	if err != nil {
		close(done)
		return err
	}
	wg.Wait()

	return nil
}