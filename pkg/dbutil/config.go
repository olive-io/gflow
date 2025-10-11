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

package dbutil

import (
	"time"

	"gorm.io/gorm/logger"
)

var (
	DefaultLevel           = "warn"
	DefaultDriver          = "sqlite"
	DefaultDBName          = "gflow.db"
	DefaultMaxIdleConns    = 10
	DefaultMaxOpenConns    = 100
	DefaultConnMaxLifetime = time.Hour
)

type Config struct {
	Level                 string        `json:"level" toml:"level"`
	Driver                string        `json:"driver" toml:"driver"`
	DataRoot              string        `json:"data_root" toml:"data_root"`
	Host                  string        `json:"host" toml:"host"`
	Username              string        `json:"username" toml:"username"`
	Password              string        `json:"password" toml:"password"`
	Port                  int32         `json:"port" toml:"port"`
	DBName                string        `json:"dbname" toml:"dbname"`
	MaxIdleConnection     int           `json:"max_idle_connection" toml:"max_idle_connections"`
	MaxOpenConnection     int           `json:"max_open_connection" toml:"max_open_connection"`
	ConnectionMaxLifetime time.Duration `json:"connection_max_lifetime" toml:"connection_max_lifetime"`
}

func NewConfig() *Config {
	cfg := &Config{
		Level:                 DefaultLevel,
		Driver:                DefaultDriver,
		DBName:                DefaultDBName,
		DataRoot:              "",
		MaxIdleConnection:     DefaultMaxIdleConns,
		MaxOpenConnection:     DefaultMaxOpenConns,
		ConnectionMaxLifetime: DefaultConnMaxLifetime,
	}

	return cfg
}

func parseLevel(level string) logger.LogLevel {
	switch level {
	case "info", "INFO":
		return logger.Info
	case "warn", "WARN":
		return logger.Warn
	case "error", "ERROR":
		return logger.Error
	default:
		return logger.Silent
	}
}
