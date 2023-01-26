/*
Copyright AppsCode Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"sigs.k8s.io/yaml"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
)

var OutputFormat string

func NewCmdCluster(f *config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cluster",
		Short:             "Manage clusters in ACE",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdCheck(f))
	cmd.AddCommand(newCmdImport(f))
	cmd.AddCommand(newCmdGet(f))
	cmd.AddCommand(newCmdConnect(f))
	cmd.AddCommand(newCmdReconfigure(f))
	cmd.AddCommand(newCmdRemove(f))

	cmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", "", "Output format (any of json,yaml,table). Default is table.")
	return cmd
}

func waitForClusterToBeReady(c *ace.Client, clusterName string) error {
	fmt.Printf("Waiting for cluster %q to be ready.....\n", clusterName)
	return wait.PollImmediate(2*time.Second, 5*time.Minute, func() (done bool, err error) {
		cluster, err := c.GetCluster(clusterName)
		if err != nil {
			return false, err
		}
		if cluster.Status.Phase == v1alpha1.ClusterPhaseActive {
			return true, nil
		}
		return false, nil
	})
}

func waitForClusterToBeRemoved(c *ace.Client, clusterName string) error {
	fmt.Printf("Waiting for cluster %q to be removed.....\n", clusterName)
	return wait.PollImmediate(2*time.Second, 5*time.Minute, func() (done bool, err error) {
		cluster, err := c.GetCluster(clusterName)
		if err != nil {
			if errors.Is(err, ace.ErrNotFound) {
				return true, nil
			}
			return false, err
		}
		if err != nil {
			return false, err
		}
		if cluster.Status.Phase == v1alpha1.ClusterPhaseInactive {
			return true, nil
		}
		return false, nil
	})
}

type clusterPrinter interface {
	printCluster(cluster *v1alpha1.ClusterInfo) error
	printClusterList(clusters []v1alpha1.ClusterInfo) error
}

func newPrinter() clusterPrinter {
	var printer clusterPrinter
	switch OutputFormat {
	case "json":
		printer = &jsonPrinter{}
	case "yaml":
		printer = &yamlPrinter{}
	default:
		printer = &tablePrinter{}
	}
	return printer
}

func printCluster(cluster *v1alpha1.ClusterInfo) error {
	printer := newPrinter()
	return printer.printCluster(cluster)
}

func printClusterList(clusters []v1alpha1.ClusterInfo) error {
	printer := newPrinter()
	return printer.printClusterList(clusters)
}

type tablePrinter struct {
}

func (p *tablePrinter) printCluster(cluster *v1alpha1.ClusterInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "NAME\tDISPLAY_NAME\tPROVIDER\tPHASE")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", cluster.Spec.Name, cluster.Spec.DisplayName, cluster.Spec.Provider, cluster.Status.Phase)
	return w.Flush()
}

func (p *tablePrinter) printClusterList(clusters []v1alpha1.ClusterInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "NAME\tDISPLAY_NAME\tPROVIDER\tPHASE")
	for i := range clusters {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", clusters[i].Spec.Name, clusters[i].Spec.DisplayName, clusters[i].Spec.Provider, clusters[i].Status.Phase)
	}
	return w.Flush()
}

type jsonPrinter struct {
}

func (p *jsonPrinter) printCluster(cluster *v1alpha1.ClusterInfo) error {
	data, err := json.MarshalIndent(cluster, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (p *jsonPrinter) printClusterList(clusters []v1alpha1.ClusterInfo) error {
	for i := range clusters {
		data, err := json.MarshalIndent(clusters[i], "", " ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

type yamlPrinter struct {
}

func (p *yamlPrinter) printCluster(cluster *v1alpha1.ClusterInfo) error {
	data, err := yaml.Marshal(cluster)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (p *yamlPrinter) printClusterList(clusters []v1alpha1.ClusterInfo) error {
	for i := range clusters {
		data, err := yaml.Marshal(clusters[i])
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}
