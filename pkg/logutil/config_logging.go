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

package logutil

import (
	"fmt"
	"net/url"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/olive-io/gflow/pkg/logutil/zapgrpc"
)

var (
	ErrLogRotationInvalidLogOutput = fmt.Errorf("--log-outputs requires a single file path when --log-rotate-config-json is defined")

	DefaultLogOutput = "default"
	JournalLogOutput = "systemd/journal"
	StdErrLogOutput  = "stderr"
	StdOutLogOutput  = "stdout"
)

type LogConfig struct {
	// Level configures log level. Only supports debug, info, warn, error, panic, or fatal. Default 'info'.
	Level string `json:"level" toml:"level"`
	// Format configures log format. Only supports json, console
	Format string `json:"format" toml:"format"`
	// LogOutputs is either:
	//  - "default" as os.Stderr,
	//  - "stderr" as os.Stderr,
	//  - "stdout" as os.Stdout,
	//  - file path to append server logs to.
	// It can be multiple when "Logger" is zap.
	Outputs []string `json:"outputs" toml:"outputs"`
	// Rotation is a passthrough allowing a log rotation JSON config to be passed directly.
	Rotation *LogRotationConfig `json:"rotation" toml:"rotation"`
	// ZapLoggerBuilder is used to build the zap logger.
	ZapLoggerBuilder func(*LogConfig) error `json:"-" toml:"-"`

	// logger logs server-side operations. The default is nil,
	// and "setupLogging" must be called before starting server.
	// Do not set logger directly.
	loggerMu *sync.RWMutex
	logger   *zap.Logger
}

// LogRotationConfig Log rotation is disabled by default.
// MaxSize:	100 // MB
// MaxAge: 0 // days (no limit)
// MaxBackups: 0 // no limit
// LocalTime: false // use computers local time, UTC by default
// Compress: false // compress the rotated log in gzip format
type LogRotationConfig struct {
	MaxSize    int  `json:"max-size" toml:"max-size"`
	MaxAge     int  `json:"max-age" toml:"max-age"`
	MaxBackups int  `json:"max-backups" toml:"max-backups"`
	LocalTime  bool `json:"localtime" toml:"localtime"`
	Compress   bool `json:"compress" toml:"compress"`
}

func NewLogConfig() LogConfig {
	return LogConfig{
		Level:    DefaultLogLevel,
		Format:   DefaultLogFormat,
		Outputs:  []string{DefaultLogOutput},
		loggerMu: new(sync.RWMutex),
		logger:   zap.NewNop(),
	}
}

// GetLogger returns the logger.
func (cfg LogConfig) GetLogger() *zap.Logger {
	cfg.loggerMu.RLock()
	l := cfg.logger
	cfg.loggerMu.RUnlock()
	return l
}

// SetupLogging initializes logging.
// Must be called after flag parsing or finishing configuring LogConfig.
func (cfg *LogConfig) SetupLogging() error {
	if len(cfg.Outputs) == 0 {
		cfg.Outputs = []string{DefaultLogOutput}
	}
	if len(cfg.Outputs) > 1 {
		for _, v := range cfg.Outputs {
			if v == DefaultLogOutput {
				return fmt.Errorf("multi logoutput for %q is not supported yet", DefaultLogOutput)
			}
		}
	}
	enableRotation := false
	if cfg.Rotation != nil {
		enableRotation = true
		if err := setupLogRotation(cfg.Outputs, cfg.Rotation); err != nil {
			return err
		}
	}

	var logFormat string
	switch cfg.Format {
	case "json":
		logFormat = "json"
	case "console", "text":
		logFormat = "console"
	default:
		logFormat = DefaultLogFormat
	}

	outputPaths, errOutputPaths := make([]string, 0), make([]string, 0)
	isJournal := false
	for _, v := range cfg.Outputs {
		switch v {
		case DefaultLogOutput:
			outputPaths = append(outputPaths, StdErrLogOutput)
			errOutputPaths = append(errOutputPaths, StdErrLogOutput)

		case JournalLogOutput:
			isJournal = true

		case StdErrLogOutput:
			outputPaths = append(outputPaths, StdErrLogOutput)
			errOutputPaths = append(errOutputPaths, StdErrLogOutput)

		case StdOutLogOutput:
			outputPaths = append(outputPaths, StdOutLogOutput)
			errOutputPaths = append(errOutputPaths, StdOutLogOutput)

		default:
			var path string
			if enableRotation {
				// append rotate scheme to logs managed by lumberjack log rotation
				if v[0:1] == "/" {
					path = fmt.Sprintf("rotate:/%%2F%s", v[1:])
				} else {
					path = fmt.Sprintf("rotate:/%s", v)
				}
			} else {
				path = v
			}
			outputPaths = append(outputPaths, path)
			errOutputPaths = append(errOutputPaths, path)
		}
	}

	if !isJournal {
		copied := DefaultZapLoggerConfig
		copied.OutputPaths = outputPaths
		copied.ErrorOutputPaths = errOutputPaths
		copied = MergeOutputPaths(copied)
		copied.Encoding = logFormat
		copied.Level = zap.NewAtomicLevelAt(ConvertToZapLevel(cfg.Level))
		if cfg.ZapLoggerBuilder == nil {
			lg, err := copied.Build(zap.AddStacktrace(zapcore.FatalLevel))
			if err != nil {
				return err
			}
			cfg.ZapLoggerBuilder = NewZapLoggerBuilder(lg)
		}
	} else {
		if len(cfg.Outputs) > 1 {
			for _, v := range cfg.Outputs {
				if v != DefaultLogOutput {
					return fmt.Errorf("running with systemd/journal but other '--log-outputs' values (%q) are configured with 'default'; override 'default' value with something else", cfg.Outputs)
				}
			}
		}

		// use stderr as fallback
		syncer, lerr := getJournalWriteSyncer()
		if lerr != nil {
			return lerr
		}

		lvl := zap.NewAtomicLevelAt(ConvertToZapLevel(cfg.Level))

		copied := DefaultZapLoggerConfig
		copied.Encoding = logFormat
		// WARN: do not change field names in encoder config
		// journald logging writer assumes field names of "level" and "caller"
		cr := zapcore.NewCore(
			zapcore.NewJSONEncoder(copied.EncoderConfig),
			syncer,
			lvl,
		)
		if cfg.ZapLoggerBuilder == nil {
			zlg := zap.New(cr,
				zap.AddCaller(),
				zap.AddStacktrace(zapcore.FatalLevel),
				zap.ErrorOutput(syncer))
			cfg.ZapLoggerBuilder = NewZapLoggerBuilder(zlg)
		}
	}

	err := cfg.ZapLoggerBuilder(cfg)
	if err != nil {
		return err
	}

	return nil
}

// NewZapLoggerBuilder generates a zap logger builder that sets given loger
func NewZapLoggerBuilder(lg *zap.Logger) func(*LogConfig) error {
	return func(cfg *LogConfig) error {
		cfg.loggerMu.Lock()
		defer cfg.loggerMu.Unlock()
		cfg.logger = lg
		return nil
	}
}

// SetupGlobalLoggers configures 'global' loggers (grpc, zapGlobal) based on the cfg.
//
// The method is not executed by embed server by default (since 3.5) to
// enable setups where grpc/zap.Global logging is configured independently
// or spans separate lifecycle (like in tests).
func (cfg *LogConfig) SetupGlobalLoggers() {
	lg := cfg.GetLogger()
	if lg != nil {
		if cfg.Level == "debug" {
			grpc.EnableTracing = true
			grpclog.SetLoggerV2(zapgrpc.NewLogger(lg, zapgrpc.WithDebug()))
		} else {
			grpclog.SetLoggerV2(zapgrpc.NewLogger(lg, zapgrpc.WithWarn()))
		}
		zap.ReplaceGlobals(lg)
	}
}

type logRotationConfig struct {
	*lumberjack.Logger
}

// Sync implements zap.Sink
func (logRotationConfig) Sync() error { return nil }

// setupLogRotation initializes log rotation for a single file path target.
func setupLogRotation(logOutputs []string, rotation *LogRotationConfig) error {
	jack := &lumberjack.Logger{
		MaxSize:    rotation.MaxSize,
		MaxAge:     rotation.MaxAge,
		MaxBackups: rotation.MaxBackups,
		LocalTime:  rotation.LocalTime,
		Compress:   rotation.Compress,
	}
	lr := logRotationConfig{Logger: jack}
	outputFilePaths := 0
	for _, v := range logOutputs {
		switch v {
		case DefaultLogOutput, StdErrLogOutput, StdOutLogOutput:
			continue
		default:
			outputFilePaths++
		}
	}
	// log rotation requires file target
	if len(logOutputs) == 1 && outputFilePaths == 0 {
		return ErrLogRotationInvalidLogOutput
	}
	// support max 1 file target for log rotation
	if outputFilePaths > 1 {
		return ErrLogRotationInvalidLogOutput
	}

	zap.RegisterSink("rotate", func(u *url.URL) (zap.Sink, error) {
		lr.Filename = u.Path[1:]
		return &lr, nil
	})
	return nil
}
