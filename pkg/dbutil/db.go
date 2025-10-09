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
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	cfg *Config
	*gorm.DB
}

func NewDB(lg *zap.Logger, cfg *Config) (*DB, error) {
	var (
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	level := parseLevel(cfg.Level)
	dbLogger := &gormLogger{
		level: level,
		lg:    lg.Sugar(),

		SlowThreshold: time.Millisecond * 200,

		IgnoreRecordNotFoundError: true,

		traceErrStr:  traceErrStr,
		traceWarnStr: traceWarnStr,
		traceStr:     traceStr,
	}

	var dialector gorm.Dialector
	switch strings.ToLower(cfg.Driver) {
	case "sqlite", "sqlite3":
		dsn := cfg.DBName
		if dataRoot := cfg.DataRoot; dataRoot != "" {
			dsn = filepath.Join(dataRoot, dsn)
		}
		if !strings.Contains(dsn, "?") {
			dsn += "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
		}
		dialector = sqlite.Open(dsn)
	case "mysql", "mariadb":
		// gorm:gorm@tcp(localhost:9910)/gorm
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		if !strings.Contains(dsn, "?") {
			dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
		}
		dialector = mysql.Open(dsn)
	case "postgres", "postgresql":
		// user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName)
		if !strings.Contains(dsn, "sslmode") {
			dsn += " sslmode=disable TimeZone=Asia/Shanghai"
		}
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %q", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 设置数据库连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if n := cfg.MaxOpenConnection; n != 0 {
		sqlDB.SetMaxIdleConns(n)
	}
	if n := cfg.MaxOpenConnection; n != 0 {
		sqlDB.SetMaxOpenConns(n)
	}
	if lifeTime := cfg.ConnectionMaxLifetime; lifeTime != 0 {
		sqlDB.SetConnMaxLifetime(lifeTime)
	}

	localDB := &DB{
		cfg: cfg,
		DB:  db,
	}

	return localDB, nil
}

func (db *DB) NewSession(ctx context.Context) *gorm.DB {
	return db.DB.Session(&gorm.Session{
		NewDB:       true,
		Initialized: true,
		Context:     ctx,
	})
}
