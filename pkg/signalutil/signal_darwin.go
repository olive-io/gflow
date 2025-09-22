package signalutil

import (
	"context"
	"os"
	"os/signal"
)

var shutdownSignals = []os.Signal{os.Interrupt}

// SetupSignalContext is same as SetupSignalHandler, but a context.Context is returned.
// Only one of SetupSignalContext and SetupSignalHandler should be called, and only can
// be called once.
func SetupSignalContext(ctx context.Context) context.Context {
	shutdownHandler = make(chan os.Signal, 1)

	ctx, cancel := context.WithCancel(ctx)
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		<-shutdownHandler
		cancel()
	}()

	return ctx
}
