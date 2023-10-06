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
	"os"

	"go.bytebuilders.dev/cli/pkg/config"
	ace "go.bytebuilders.dev/client"

	"github.com/spf13/cobra"
	"kubeops.dev/installer/apis/installer/v1alpha1"
)

func newCmdLogin() *cobra.Command {
	cred := v1alpha1.BasicAuth{}
	cmd := &cobra.Command{
		Use:               "login",
		Short:             "Establish a authenticated session with the api endpoint",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := login(cred)
			if err != nil {
				return err
			}
			fmt.Println("Successfully logged in")
			return nil
		},
	}

	cmd.Flags().StringVar(&cred.Username, "username", "", "Name of user to login")
	cmd.Flags().StringVar(&cred.Password, "password", "", "Password to use to log in")

	cmd.Flags().StringVar(&AccessToken, "access-token", "", "Access token to call API")

	return cmd
}

func login(cred v1alpha1.BasicAuth) error {
	ctx, err := config.GetContext()
	if err != nil {
		return err
	}
	client := ace.NewClient(ctx.Endpoint)

	if AccessToken != "" {
		if err = SetToken(AccessToken); err != nil {
			return err
		}
		return nil
	}

	if cred.Username == "" {
		cred.Username = os.Getenv(BB_USERNAME)
	}
	if cred.Password == "" {
		cred.Password = os.Getenv(BB_PASSWORD)
	}
	if cred.Username == "" || cred.Password == "" {
		return fmt.Errorf("missing credentials. Please provide both username and password")
	}
	cookies, err := client.Signin(ace.SignInParams{UserName: cred.Username, Password: cred.Password})
	if err != nil {
		return err
	}
	ctx.Cookies = make([]http.Cookie, 0)
	for i := range cookies {
		if cookies[i].Name == csrfCookie || cookies[i].Name == sessionCookie {
			ctx.Cookies = append(ctx.Cookies, cookies[i])
		}
	}
	return config.SetContext(*ctx)
}
