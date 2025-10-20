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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	"github.com/olive-io/gflow/pkg/dbutil"
	"github.com/olive-io/gflow/pkg/logutil"
)

var (
	DefaultListenAddr = "localhost:6550"
)

type TLS struct {
	CaFile   string `json:"ca_file" toml:"ca_file"`
	CertFile string `json:"cert_file" toml:"cert_file"`
	KeyFile  string `json:"key_file" toml:"key_file"`
}

type ServerConfig struct {
	ID     string `json:"id" toml:"id"`
	Listen string `json:"listen" toml:"listen"`
	TLS    *TLS   `json:"tls" toml:"tls"`
}

type MessageQueueConfig struct {
	Name     string          `json:"name" toml:"name"`
	RabbitMQ *RabbitMQConfig `json:"rabbitmq" toml:"rabbitmq"`
}

type RabbitMQConfig struct {
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
	// rabbit host and port (localhost:5672)
	Host string `json:"host" toml:"host"`
}

type EmailConfig struct {
	Name string `json:"name" toml:"name"`
	// email server and port
	Host     string `json:"host" toml:"host"`
	Username string `json:"username" toml:"username"`
	Secret   string `json:"secret" toml:"secret"`
}

type PluginConfig struct {
	SendTask    *SendTaskPluginConfig    `json:"sendTask" toml:"sendTask"`
	ReceiveTask *ReceiveTaskPluginConfig `json:"receiveTask" toml:"receiveTask"`
	ScriptTask  *ScriptTaskPluginConfig  `json:"scriptTask" toml:"scriptTask"`
}

type RabbitMQConfigWithRef struct {
	*RabbitMQConfig `json:",inline" toml:",inline"`

	Ref string `json:"ref" toml:"ref"`
}

type SendTaskPluginConfig struct {
	RabbitMQ *RabbitMQConfigWithRef `json:"rabbitmq" toml:"rabbitmq"`
}

type ReceiveTaskPluginConfig struct {
	RabbitMQ *RabbitMQConfigWithRef `json:"rabbitmq" toml:"rabbitmq"`
}

type ScriptTaskPluginConfig struct {
}

type Config struct {
	once sync.Once

	Server   ServerConfig   `json:"server" toml:"server"`
	Database *dbutil.Config `json:"database" toml:"database"`

	Log *logutil.LogConfig `json:"log" toml:"log"`

	Plugin *PluginConfig `json:"plugin" toml:"plugin"`

	MQ []MessageQueueConfig `json:"mq" toml:"mq"`

	Email []EmailConfig `json:"email" toml:"email"`
}

func NewConfig() *Config {
	lc := logutil.NewLogConfig()
	dbCfg := dbutil.NewConfig()
	cfg := &Config{
		Server: ServerConfig{
			Listen: DefaultListenAddr,
		},
		Database: dbCfg,
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
	if cfg.Database == nil {
		return errors.New("database config is required")
	}

	err := cfg.Log.SetupLogging()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	cfg.Log.SetupGlobalLoggers()

	if cfg.Server.ID == "" {
		return errors.New("server id is required")
	}

	if cfg.Database.DataRoot != "" {
		_, err := os.Stat(cfg.Database.DataRoot)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("read data root directory: %w", err)
			}
			_ = os.MkdirAll(cfg.Database.DataRoot, 0755)
		}
		if strings.HasPrefix(cfg.Database.DataRoot, "~") || strings.HasPrefix(cfg.Database.DataRoot, "./") {
			abs, err := filepath.Abs(cfg.Database.DataRoot)
			if err != nil {
				return fmt.Errorf("get data directory abs path: %w", err)
			}
			cfg.Database.DataRoot = abs
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

func (cfg *Config) FormatPluginConfig() (*PluginConfig, error) {
	if cfg.Plugin == nil || cfg.Plugin.SendTask == nil {
		return nil, fmt.Errorf("send task plugin config is required")
	}

	sc := cfg.Plugin.SendTask
	if sc != nil {
		if mq := sc.RabbitMQ; mq != nil {
			if mq.Ref != "" && mq.RabbitMQConfig == nil {
				ref := mq.Ref
				exists := false
				for _, item := range cfg.MQ {
					if item.Name == ref {
						exists = true
						mq.RabbitMQConfig = item.RabbitMQ
					}
				}
				if !exists {
					return nil, fmt.Errorf("rabbitmq reference not found: %s", ref)
				}
			}
		}
	}

	return cfg.Plugin, nil
}

func (cfg *Config) Logger() *zap.Logger {
	return cfg.Log.GetLogger()
}
