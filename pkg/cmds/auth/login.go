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

	"go.bytebuilders.dev/ace/pkg/config"
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

	cmd.Flags().StringVar(&AccessToken, "token", os.Getenv(ACE_TOKEN), "Access token to call API")

	cmd.Flags().StringVar(&cred.Username, "username", os.Getenv(ACE_USERNAME), "Name of user to login")
	cmd.Flags().StringVar(&cred.Password, "password", os.Getenv(ACE_PASSWORD), "Password to use to log in")

	return cmd
}

func login(cred v1alpha1.BasicAuth) error {
	ctx, err := config.GetContext()
	if err != nil {
		return err
	}

	if AccessToken != "" {
		ctx.Token = AccessToken
		return config.SetContext(*ctx)
	}

	if cred.Username == "" || cred.Password == "" {
		return fmt.Errorf("missing credentials. Please provide both username and password")
	}
	client := ace.NewClient(ctx.Endpoint)
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
