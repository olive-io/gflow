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

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	"github.com/olive-io/gflow/pkg/logutil"
)

var (
	DefaultDataRoot   = "./data"
	DefaultListenAddr = "localhost:6550"
)

type TLS struct {
	CaFile   string `json:"ca-file" toml:"ca-file"`
	CertFile string `json:"cert-file" toml:"cert-file"`
	KeyFile  string `json:"key-file" toml:"key-file"`
}

type Config struct {
	once sync.Once

	Listen string `json:"listen" toml:"listen"`

	TLS *TLS `json:"tls" toml:"tls"`

	DataRoot string `json:"data-root" toml:"data-root"`

	Log *logutil.LogConfig `json:"log" toml:"log"`
}

func NewConfig() *Config {
	lc := logutil.NewLogConfig()
	cfg := &Config{
		Listen:   DefaultListenAddr,
		DataRoot: DefaultDataRoot,
		Log:      &lc,
	}

	return cfg
}

func (cfg *Config) Init() error {
	var err error
	cfg.once.Do(func() {
		err = cfg.init()
	})
	return err
}

func (cfg *Config) init() error {
	if cfg.Log == nil {
		lc := logutil.NewLogConfig()
		cfg.Log = &lc
	}
	err := cfg.Log.SetupLogging()
	cfg.Log.SetupGlobalLoggers()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if cfg.DataRoot == "" {
		home, _ := os.UserHomeDir()
		cfg.DataRoot = filepath.Join(home, ".olive")
		_ = os.MkdirAll(cfg.DataRoot, 0755)
	} else {
		_, err := os.Stat(cfg.DataRoot)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("read data root directory: %w", err)
			}
			_ = os.MkdirAll(cfg.DataRoot, 0755)
		}
		if strings.HasPrefix(cfg.DataRoot, "~") || strings.HasPrefix(cfg.DataRoot, "./") {
			abs, err := filepath.Abs(cfg.DataRoot)
			if err != nil {
				return fmt.Errorf("get data directory abs path: %w", err)
			}
			cfg.DataRoot = abs
		}
	}
	return nil
}

func FromPath(filename string) (*Config, error) {
	var cfg Config
	var err error
	ext := filepath.Ext(filename)
	switch ext {
	case ".toml":
		_, err = toml.DecodeFile(filename, &cfg)
	case ".yaml", ".yml":
		err = yaml.Unmarshal([]byte(filename), &cfg)
	case ".json":
		err = json.Unmarshal([]byte(filename), &cfg)
	default:
		return nil, fmt.Errorf("invalid config format: %s", ext)
	}
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves config text to specific file path
func (cfg *Config) Save(filename string) error {
	var err error
	var data []byte
	ext := filepath.Ext(filename)
	switch ext {
	case ".toml":
		buf := bytes.NewBufferString("")
		err = toml.NewEncoder(buf).Encode(cfg)
		if err == nil {
			data = buf.Bytes()
		}
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
	case ".json":
		data, err = json.Marshal(cfg)
	default:
		return fmt.Errorf("invalid config format: %s", ext)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0755)
}

func (cfg *Config) Logger() *zap.Logger {
	return cfg.Log.GetLogger()
}
