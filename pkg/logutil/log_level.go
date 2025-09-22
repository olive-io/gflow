package logutil

import (
	"go.uber.org/zap/zapcore"
)

var DefaultLogLevel = "info"

// ConvertToZapLevel converts log level string to zapcore.Level.
func ConvertToZapLevel(lvl string) zapcore.Level {
	var level zapcore.Level
	if err := level.Set(lvl); err != nil {
		panic(err)
	}
	return level
}
