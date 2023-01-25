package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/config"
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
