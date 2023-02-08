package auth

import (
	"fmt"
	"net/http"

	"go.bytebuilders.dev/ace-cli/pkg/config"
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
