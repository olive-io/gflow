//go:build !windows
// +build !windows

package logutil

import (
	"fmt"
	"os"

	"go.uber.org/zap/zapcore"
)

var DefaultLineEnding = zapcore.DefaultLineEnding

// use stderr as fallback
func getJournalWriteSyncer() (zapcore.WriteSyncer, error) {
	jw, err := NewJournalWriter(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("can't find journal (%v)", err)
	}
	return zapcore.AddSync(jw), nil
}
