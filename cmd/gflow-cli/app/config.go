/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/olive-io/gflow/clientgo"
)

const (
	DefaultEndpoint = "localhost:6550"
	ConfigFileName  = "gfcli.yaml"
	ConfigDirName   = ".gflow"
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ConfigFileName
	}
	return filepath.Join(home, ConfigDirName, ConfigFileName)
}

func GetConfig(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("endpoint", DefaultEndpoint)
	v.SetDefault("token", "")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("gfcli")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ConfigDirName))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	v := viper.New()

	v.Set("endpoint", cfg.Endpoint)
	v.Set("token", cfg.Token)

	if path == "" {
		path = DefaultConfigPath()
	}

	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return v.WriteConfigAs(path)
}

func NewClient(cfg *Config) (*clientgo.Client, error) {
	clientCfg := clientgo.NewConfig(cfg.Endpoint)
	clientCfg.Token = cfg.Token

	return clientgo.NewClient(clientCfg)
}
