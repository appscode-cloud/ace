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

package auth

import (
	"os"

	"github.com/spf13/cobra"
	"kubeops.dev/installer/apis/installer/v1alpha1"
)

func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "auth",
		Short:             "Manage authentication for the CLI",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(newCmdLogin())
	cmd.AddCommand(newCmdLogout())

	return cmd
}

const (
	BB_USERNAME     = "BB_USERNAME"
	BB_PASSWORD     = "BB_PASSWORD"
	BB_ACCESS_TOKEN = "BB_ACCESS_TOKEN"

	csrfCookie    = "_csrf"
	sessionCookie = "i_like_bytebuilders"
)

var AccessToken string

func GetBasicAuthCredFromEnv() *v1alpha1.BasicAuth {
	user := os.Getenv(BB_USERNAME)
	password := os.Getenv(BB_PASSWORD)
	if user == "" || password == "" {
		return nil
	}
	return &v1alpha1.BasicAuth{
		Username: user,
		Password: password,
	}
}

func GetAuthTokenFromEnv() string {
	return os.Getenv(BB_ACCESS_TOKEN)
}
