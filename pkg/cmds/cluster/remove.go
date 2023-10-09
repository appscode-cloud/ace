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

	"go.bytebuilders.dev/cli/pkg/config"
	"go.bytebuilders.dev/cli/pkg/printer"
	ace "go.bytebuilders.dev/client"
	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func newCmdRemove(f *config.Factory) *cobra.Command {
	opts := clustermodel.RemovalOptions{}
	cmd := &cobra.Command{
		Use:               "remove",
		Short:             "Remove a cluster from ACE platform",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Components.AllFeatures {
				opts.Components.FeatureSets = defaultFeatureSet
			}
			err := removeCluster(f, opts)
			if err != nil {
				if errors.Is(err, ace.ErrNotFound) {
					fmt.Println("Cluster has been removed already.")
					return nil
				}
				return fmt.Errorf("failed to remove cluster. Reason: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the cluster to get")
	cmd.Flags().BoolVar(&opts.Components.FluxCD, "remove-fluxcd", true, "Specify whether to remove FluxCD or not (default true).")
	cmd.Flags().BoolVar(&opts.Components.AllFeatures, "all-features", false, "Remove all features")
	return cmd
}

func removeCluster(f *config.Factory, opts clustermodel.RemovalOptions) error {
	fmt.Println("Removing cluster......")
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
			fmt.Println("Failed to log removal steps. Reason: ", err)
		}
	}()

	err = c.RemoveCluster(opts, responseID)
	if err != nil {
		close(done)
		return err
	}
	wg.Wait()

	return nil
}
