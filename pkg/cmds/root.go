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

package cmds

import (
	"github.com/spf13/cobra"
	"go.bytebuilders.dev/ace-cli/pkg/cmds/auth"
	"go.bytebuilders.dev/ace-cli/pkg/cmds/cluster"
	cmdconfig "go.bytebuilders.dev/ace-cli/pkg/cmds/config"
	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	v "gomodules.xyz/x/version"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "ace",
		Short:             `CLI to interact with ACE platform`,
		Long:              `A cli to interact with ACE (AppsCode Container Engine) platform`,
		DisableAutoGenTag: true,
	}
	rootCmd.PersistentFlags().StringVar(&config.CurrentContext, "context", "", "Use this as current context instead of one from configuration file")

	f := &config.Factory{
		Client: aceClient,
	}
	rootCmd.AddCommand(cmdconfig.NewCmdConfig())
	rootCmd.AddCommand(cluster.NewCmdCluster(f))
	rootCmd.AddCommand(auth.NewCmdAuth())

	rootCmd.AddCommand(v.NewCmdVersion())
	rootCmd.AddCommand(NewCmdCompletion())

	return rootCmd
}

func aceClient() (*ace.Client, error) {
	cfg, err := config.GetContext()
	if err != nil {
		return nil, err
	}
	client := ace.NewClient(cfg.Endpoint)

	if cred := auth.GetBasicAuthCredFromEnv(); cred != nil {
		client = client.WithBasicAuth(cred.Username, cred.Password)
	}
	if cfg.Cookies != nil {
		client = client.WithCookies(cfg.Cookies)
	}

	return client, err
}
