package config

type Config struct {
	Version        string    `json:"version,omitempty"`
	CurrentContext string    `json:"currentContext,omitempty"`
	Contexts       []Context `json:"contexts,omitempty"`
}

type Context struct {
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint,omitempty"`
	Token     string    `json:"token,omitempty"`
	BasicAuth BasicAuth `json:"basicAuth,omitempty"`
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoadConfig() (*Context, error) {

	return &Context{
		Endpoint: "http://api.bb.test:3003",
		BasicAuth: BasicAuth{
			Username: "appscode",
			Password: "password",
		},
	}, nil
}
