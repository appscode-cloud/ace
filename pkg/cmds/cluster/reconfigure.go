/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"errors"
	"fmt"
	"sync"

	"go.bytebuilders.dev/ace/pkg/config"
	"go.bytebuilders.dev/ace/pkg/printer"
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
			if !opts.Components.AllFeatures {
				opts.Components.FeatureSets = defaultFeatureSet
			}
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
	cmd.Flags().StringVar(&opts.BasicInfo.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "install-fluxcd", true, "Specify whether to install FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.AllFeatures, "all-features", false, "Install all features")
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
