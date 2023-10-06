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

package config

import (
	"fmt"

	"go.bytebuilders.dev/cli/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newCmdView() *cobra.Command {
	var showSensitiveData bool
	cmd := &cobra.Command{
		Use:               "view",
		Short:             "View current CLI configurations",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewConfig(showSensitiveData)
		},
	}
	cmd.Flags().BoolVar(&showSensitiveData, "show-sensitive-data", false, "Show sensitive data (default false)")
	return cmd
}

func viewConfig(showSensitiveData bool) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}
	if !showSensitiveData {
		cfg.MaskSensitiveData()
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
