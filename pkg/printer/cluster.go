package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"

	"sigs.k8s.io/yaml"
)

var OutputFormat string

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

func PrintCluster(cluster *v1alpha1.ClusterInfo) error {
	printer := newPrinter()
	return printer.printCluster(cluster)
}

func PrintClusterList(clusters []v1alpha1.ClusterInfo) error {
	printer := newPrinter()
	return printer.printClusterList(clusters)
}

type tablePrinter struct{}

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

type jsonPrinter struct{}

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

type yamlPrinter struct{}

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
