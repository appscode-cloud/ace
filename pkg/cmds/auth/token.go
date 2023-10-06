package auth

import "go.bytebuilders.dev/cli/pkg/config"

func SetToken(token string) error {
	ctx, err := config.GetContext()
	if err != nil {
		return err
	}

	ctx.Token = token
	return config.SetContext(*ctx)
}
