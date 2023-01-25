package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

const (
	configVersion = "v1"

	BB_USERNAME = "BB_USERNAME"
	BB_PASSWORD = "BB_PASSWORD"
)

var CurrentContext string

type Config struct {
	Version        string    `yaml:"version,omitempty"`
	CurrentContext string    `yaml:"currentContext,omitempty"`
	Contexts       []Context `yaml:"contexts,omitempty"`
}

type Context struct {
	Name     string       `yaml:"name"`
	Endpoint string       `yaml:"endpoint,omitempty"`
	Token    string       `yaml:"token,omitempty"`
	Cookie   *http.Cookie `yaml:"cookie,omitempty"`
}

func ReadConfig() (Config, error) {
	configFile, err := getConfigFilepath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultConfig(), nil
		}
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func GetContext() (*Context, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	curContext := config.CurrentContext
	if CurrentContext != "" {
		curContext = CurrentContext
	}
	for i := range config.Contexts {
		if config.Contexts[i].Name == curContext {
			return &config.Contexts[i], nil
		}
	}

	return nil, fmt.Errorf("no data found for context: %s", curContext)
}

func SetContext(ctx Context) error {
	config, err := ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		config = defaultConfig()
	}
	contextExist := false
	for i := range config.Contexts {
		if config.Contexts[i].Name == ctx.Name {
			contextExist = true
			config.Contexts[i] = updateContext(config.Contexts[i], ctx)
		}
	}
	if !contextExist {
		config.Contexts = append(config.Contexts, ctx)
	}
	return config.save()
}

func (cfg *Config) MaskSensitiveData() {
	for i := range cfg.Contexts {
		cfg.Contexts[i].Cookie = nil
	}
}
func (cfg *Config) save() error {
	configFile, err := getConfigFilepath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configFile); err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(configFile, 0664)
		if err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0664)
}

func getConfigFilepath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "ace", fmt.Sprintf("config_%s.yaml", configVersion)), nil
}

func defaultConfig() Config {
	return Config{
		Version: configVersion,
		Contexts: []Context{
			{
				Name:     "bytebuilders",
				Endpoint: "https://api.byte.builders",
			},
			{
				Name:     "appscodeninja",
				Endpoint: "https://api.appscode.ninja",
			},
		},
		CurrentContext: "bytebuilders",
	}
}

func updateContext(cur, new Context) Context {
	if new.Endpoint != "" {
		cur.Endpoint = new.Endpoint
	}
	if new.Token != "" {
		cur.Token = new.Token
	}
	if new.Cookie != nil {
		cur.Cookie = new.Cookie
	}
	return cur
}
