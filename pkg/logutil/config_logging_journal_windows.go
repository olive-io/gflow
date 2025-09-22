//go:build windows
// +build windows

package logutil

import (
	"os"

	"go.uber.org/zap/zapcore"
)

var DefaultLineEnding = "\r\n"

func getJournalWriteSyncer() (zapcore.WriteSyncer, error) {
	return zapcore.AddSync(os.Stderr), nil
}
