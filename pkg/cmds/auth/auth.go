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
	BB_USERNAME = "BB_USERNAME"
	BB_PASSWORD = "BB_PASSWORD"

	csrfCookie    = "_csrf"
	sessionCookie = "i_like_bytebuilders"
)

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
