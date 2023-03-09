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
)

var (
	CurrentContext string
	Organization   string
)

var ErrContextNotFound = errors.New("context does not exist")

type Config struct {
	Version        string    `yaml:"version,omitempty"`
	CurrentContext string    `yaml:"currentContext,omitempty"`
	Contexts       []Context `yaml:"contexts,omitempty"`
}

type Context struct {
	Name     string        `yaml:"name"`
	Endpoint string        `yaml:"endpoint,omitempty"`
	Token    string        `yaml:"token,omitempty"`
	Cookies  []http.Cookie `yaml:"cookies,omitempty"`
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

	curContext := config.getCurrentContext()
	for i := range config.Contexts {
		if config.Contexts[i].Name == curContext {
			return &config.Contexts[i], nil
		}
	}

	return nil, fmt.Errorf("no data found for context: %s", curContext)
}

func SetContext(ctx Context) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	contextExist, idx := cfg.isContextExist(ctx.Name)

	if contextExist {
		cfg.Contexts[idx] = ctx
	} else {
		cfg.Contexts = append(cfg.Contexts, ctx)
	}
	cfg.CurrentContext = ctx.Name
	return cfg.save()
}

func DeleteContext(ctx string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if ctx == cfg.CurrentContext {
		return fmt.Errorf("can't delete the context. Reason: %q is set as current context", ctx)
	}
	contextExist, idx := cfg.isContextExist(ctx)
	if !contextExist {
		return ErrContextNotFound
	}
	length := len(cfg.Contexts)
	cfg.Contexts[idx] = cfg.Contexts[length-1]
	cfg.Contexts = cfg.Contexts[:length-1]
	return cfg.save()
}

func SetCurrentContext(ctx string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	contextExist, _ := cfg.isContextExist(ctx)
	if !contextExist {
		return ErrContextNotFound
	}
	cfg.CurrentContext = ctx
	return cfg.save()
}

func (cfg *Config) MaskSensitiveData() {
	for i := range cfg.Contexts {
		cfg.Contexts[i].Cookies = nil
		if cfg.Contexts[i].Token != "" {
			cfg.Contexts[i].Token = "<REDACTED>"
		}
	}
}

func (c *Config) getCurrentContext() string {
	if CurrentContext != "" {
		return CurrentContext
	}
	return c.CurrentContext
}

func (cfg *Config) isContextExist(ctx string) (bool, int) {
	for i := range cfg.Contexts {
		if cfg.Contexts[i].Name == ctx {
			return true, i
		}
	}
	return false, -1
}

func (cfg *Config) save() error {
	configFile, err := getConfigFilepath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configFile); err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(filepath.Dir(configFile), 0o700)
		if err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0o700)
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
