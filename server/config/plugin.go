/*
Copyright 2026 The gflow Authors

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
	"fmt"
)

type MessageQueueConfig struct {
	Name     string          `mapstructure:"name" json:"name" toml:"name"`
	RabbitMQ *RabbitMQConfig `mapstructure:"rabbitmq" json:"rabbitmq" toml:"rabbitmq"`
}

type RabbitMQConfig struct {
	Username string `mapstructure:"username" json:"username" toml:"username"`
	Password string `mapstructure:"password" json:"password" toml:"password"`
	// rabbit host and port (localhost:5672)
	Host string `mapstructure:"host" json:"host" toml:"host"`
}

type ExecutorConfig struct {
	Name  string       `mapstructure:"name" json:"name" toml:"name"`
	Shell *ShellConfig `mapstructure:"shell" json:"shell" toml:"shell"`
}

type ShellConfig struct {
	Shell string   `mapstructure:"shell" json:"shell" toml:"shell"`
	Args  []string `mapstructure:"args" json:"args" toml:"args"`
	Env   []string `mapstructure:"env" json:"env" toml:"env"`
}

type EmailConfig struct {
	Name string `mapstructure:"name" json:"name" toml:"name"`
	// email server and port
	Host     string `mapstructure:"host" json:"host" toml:"host"`
	Username string `mapstructure:"username" json:"username" toml:"username"`
	Secret   string `mapstructure:"secret" json:"secret" toml:"secret"`
}

type PluginConfig struct {
	SendTask    *SendTaskPluginConfig    `mapstructure:"sendTask" json:"sendTask" toml:"sendTask"`
	ReceiveTask *ReceiveTaskPluginConfig `mapstructure:"receiveTask" json:"receiveTask" toml:"receiveTask"`
	ScriptTask  *ScriptTaskPluginConfig  `mapstructure:"scriptTask" json:"scriptTask" toml:"scriptTask"`
}

type RabbitMQConfigWithRef struct {
	*RabbitMQConfig `mapstructure:",squash" json:",inline" toml:",inline"`

	Ref string `mapstructure:"ref" json:"ref" toml:"ref"`
}

type ShellScriptConfigWithRef struct {
	*ShellConfig `mapstructure:",squash" json:",inline" toml:",inline"`

	Ref string `mapstructure:"ref" json:"ref" toml:"ref"`
}

type SendTaskPluginConfig struct {
	RabbitMQ *RabbitMQConfigWithRef `mapstructure:"rabbitmq" json:"rabbitmq" toml:"rabbitmq"`
}

type ReceiveTaskPluginConfig struct {
	RabbitMQ *RabbitMQConfigWithRef `mapstructure:"rabbitmq" json:"rabbitmq" toml:"rabbitmq"`
}

type ScriptTaskPluginConfig struct {
	Timeout int32                     `mapstructure:"timeout" json:"timeout" toml:"timeout"`
	Shell   *ShellScriptConfigWithRef `mapstructure:"shell" json:"shell" toml:"shell"`
}

func (cfg *Config) FormatPluginConfig() (*PluginConfig, error) {
	if cfg.Plugin == nil {
		return nil, fmt.Errorf("plugin config is required")
	}

	if sc := cfg.Plugin.SendTask; sc != nil {
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
					return nil, fmt.Errorf("send task rabbitmq reference not found: %s", ref)
				}
			}
		}
	}
	if rc := cfg.Plugin.ReceiveTask; rc != nil {
		if mq := rc.RabbitMQ; mq != nil {
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
					return nil, fmt.Errorf("receive task rabbitmq reference not found: %s", ref)
				}
			}
		}
	}

	if st := cfg.Plugin.ScriptTask; st != nil {
		if shell := st.Shell; shell != nil {
			if shell.Ref != "" && shell.ShellConfig == nil {
				ref := shell.Ref
				exists := false
				for _, item := range cfg.Executor {
					if item.Name == ref {
						exists = true
						shell.ShellConfig = item.Shell
					}
				}
				if !exists {
					return nil, fmt.Errorf("shell reference not found: %s", ref)
				}
			}
		}
	}

	return cfg.Plugin, nil
}
