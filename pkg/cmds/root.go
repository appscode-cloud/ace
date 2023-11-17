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
	"os"
	"os/signal"
	"syscall"

	"go.bytebuilders.dev/ace/pkg/cmds/auth"
	"go.bytebuilders.dev/ace/pkg/cmds/cloud_swap"
	"go.bytebuilders.dev/ace/pkg/cmds/cluster"
	cmdconfig "go.bytebuilders.dev/ace/pkg/cmds/config"
	"go.bytebuilders.dev/ace/pkg/config"
	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
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
	rootCmd.PersistentFlags().StringVar(&config.Organization, "org", "", "Use this organization for instead of auto-detecting current one")

	f := &config.Factory{
		Client:    aceClient,
		Canceller: canceller,
	}
	rootCmd.AddCommand(cmdconfig.NewCmdConfig())
	rootCmd.AddCommand(cluster.NewCmdCluster(f))
	rootCmd.AddCommand(auth.NewCmdAuth())

	rootCmd.AddCommand(cloud_swap.NewCmdCloudSwap())

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
	if config.Organization != "" {
		client = client.WithOrganization(config.Organization)
	}

	if cfg.Token != "" {
		client = client.WithAccessToken(cfg.Token)
	}

	if cred := auth.GetBasicAuthCredFromEnv(); cred != nil {
		client = client.WithBasicAuth(cred.Username, cred.Password)
	}
	if token := auth.GetAuthTokenFromEnv(); token != "" {
		client = client.WithAccessToken(token)
	}

	if cfg.Cookies != nil {
		client = client.WithCookies(cfg.Cookies)
	}

	return client, err
}

func canceller() chan os.Signal {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	return stopCh
}
