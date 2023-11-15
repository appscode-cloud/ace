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
	"fmt"
	"net/http"

	"go.bytebuilders.dev/ace/pkg/config"
	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
)

func newCmdLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "logout",
		Short:             "End current authenticated session with the api endpoint",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := logout()
			if err != nil {
				return err
			}
			fmt.Println("Successfully logged out")
			return nil
		},
	}

	return cmd
}

func logout() error {
	ctx, err := config.GetContext()
	if err != nil {
		return err
	}
	client := ace.NewClient(ctx.Endpoint).WithCookies(ctx.Cookies)

	err = client.Signout()
	if err != nil {
		return err
	}
	ctx.Cookies = []http.Cookie{}
	return config.SetContext(*ctx)
}
