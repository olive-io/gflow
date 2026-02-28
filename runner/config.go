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

package runner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	json "github.com/bytedance/sonic"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"sigs.k8s.io/yaml"

	"github.com/olive-io/gflow/pkg/logutil"
	traceutil "github.com/olive-io/gflow/pkg/trace"
)

var (
	DefaultID                = "runner"
	DefaultTargets           = "localhost:6550"
	DefaultHeartBeatInterval = time.Second * 30
)

type Config struct {
	once sync.Once

	ID      string `mapstructure:"id" json:"id" toml:"id"`
	Targets string `mapstructure:"targets" json:"targets" toml:"targets"`

	CertFile   string `mapstructure:"cert_file" json:"cert_file" toml:"cert_file"`
	KeyFile    string `mapstructure:"key_file" json:"key_file" toml:"key_file"`
	CaFile     string `mapstructure:"ca_file" json:"ca_file" toml:"ca_file"`
	ServerName string `mapstructure:"server_name" json:"server_name" toml:"server_name"`

	HeartBeatInterval time.Duration `mapstructure:"heartbeat_interval" json:"heartbeat_interval" toml:"heartbeat_interval"`

	Metadata map[string]string `mapstructure:"metadata" json:"metadata" toml:"metadata"`

	Log *logutil.LogConfig `mapstructure:"log" json:"log" toml:"log"`

	Trace *traceutil.Config `mapstructure:"trace" json:"trace" toml:"trace"`
}

func NewConfig() *Config {
	logCfg := logutil.NewLogConfig()
	cfg := &Config{
		ID:                DefaultID,
		Targets:           DefaultTargets,
		HeartBeatInterval: DefaultHeartBeatInterval,
		Metadata:          map[string]string{},

		Log: &logCfg,
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
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	cfg.Log.SetupGlobalLoggers()

	if cfg.ID == "" {
		return fmt.Errorf("id is required")
	}
	if len(cfg.Targets) == 0 {
		return fmt.Errorf("targets is required")
	}

	return nil
}

func FromConfigPath(filename string) (*Config, error) {
	var cfg Config

	v := viper.New()
	v.SetDefault("id", DefaultID)
	v.SetDefault("targets", DefaultTargets)
	v.SetDefault("heartbeat_interval", DefaultHeartBeatInterval)
	v.SetConfigFile(filename)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %q: %w", filename, err)
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", filename, err)
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

func (cfg *Config) TargetURLs() []string {
	parts := strings.Split(cfg.Targets, ",")
	targets := make([]string, len(parts))
	for i, part := range parts {
		targets[i] = part
	}
	return targets
}

func (cfg *Config) Logger() *otelzap.Logger {
	return cfg.Log.GetLogger()
}
