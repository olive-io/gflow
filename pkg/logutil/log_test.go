package logutil

import (
	"testing"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	lc := NewLogConfig()
	lc.Format = "console"
	err := lc.SetupLogging()
	if err != nil {
		t.Fatal(err)
	}

	lc.GetLogger().With(zap.String("a", "b")).Debug("info message")
	lc.GetLogger().With(zap.String("a", "b")).Info("info message")
	lc.GetLogger().With(zap.String("a", "b")).Warn("info message")
	lc.GetLogger().With(zap.String("a", "b")).Error("info message")
}
