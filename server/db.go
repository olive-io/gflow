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

package server

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/olive-io/gflow/server/config"
)

type gormLogger struct {
	level logger.LogLevel
	lg    *zap.SugaredLogger

	SlowThreshold time.Duration

	IgnoreRecordNotFoundError bool

	traceStr, traceErrStr, traceWarnStr string
}

func (glog *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	glog.level = level
	return glog
}

func (glog gormLogger) Info(ctx context.Context, format string, i ...interface{}) {
	if glog.level >= logger.Info {
		glog.lg.Infof(format, i...)
	}
}

func (glog gormLogger) Warn(ctx context.Context, format string, i ...interface{}) {
	if glog.level >= logger.Warn {
		glog.lg.Warnf(format, i...)
	}
}

func (glog gormLogger) Error(ctx context.Context, format string, i ...interface{}) {
	if glog.level >= logger.Error {
		glog.lg.Errorf(format, i...)
	}
}

func (glog gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if glog.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && glog.level >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !glog.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			glog.lg.Errorf(glog.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.lg.Errorf(glog.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > glog.SlowThreshold && glog.SlowThreshold != 0 && glog.level >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", glog.SlowThreshold)
		if rows == -1 {
			glog.lg.Logf(zap.WarnLevel, glog.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.lg.Logf(zap.WarnLevel, glog.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case glog.level == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			glog.lg.Infof(glog.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.lg.Infof(glog.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func openLocalDB(lg *zap.Logger, dbCfg *config.DatabaseConfig) (*gorm.DB, error) {
	var (
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	dbLogger := &gormLogger{
		level: logger.Warn,
		lg:    lg.Sugar(),

		SlowThreshold: time.Millisecond * 200,

		IgnoreRecordNotFoundError: true,

		traceErrStr:  traceErrStr,
		traceWarnStr: traceWarnStr,
		traceStr:     traceStr,
	}

	dataDir := dbCfg.DataRoot
	dbPath := filepath.Join(dataDir, "gflow.db")
	dbPath += "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, err
	}
	// 设置数据库连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
