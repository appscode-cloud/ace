package auth

import (
	"fmt"
	"os"

	"go.bytebuilders.dev/ace-cli/pkg/config"
	ace "go.bytebuilders.dev/client"
	"kubeops.dev/installer/apis/installer/v1alpha1"

	"github.com/spf13/cobra"
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

	return cmd
}

func login(cred v1alpha1.BasicAuth) error {
	ctx, err := config.GetContext()
	if err != nil {
		return err
	}
	client := ace.NewClient(ctx.Endpoint)

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
	ctx.Cookies = cookies
	return config.SetContext(*ctx)
}
