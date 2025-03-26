package catapp

import (
	"log/slog"
	"os"
	"time"

	"github.com/surkovvs/gocat/catapp/interfaces"
)

type appOption func(*app)

func WithName(name string) appOption {
	return func(a *app) {
		a.name = name
	}
}

func WithLogger(logger interfaces.Logger) appOption {
	return func(a *app) {
		a.logger = logger
	}
}

func WithInitTimeout(to time.Duration) appOption {
	return func(a *app) {
		a.execution.initTimeout = &to
	}
}

func WithProvidedSigs(sigs ...os.Signal) appOption {
	return func(a *app) {
		a.shutdown.sigs = sigs
	}
}

func WithShutdownTimeout(to time.Duration) appOption {
	return func(a *app) {
		a.shutdown.timeout = &to
	}
}

type logWrap struct {
	logger interfaces.Logger
}

func newLogWrap(logger interfaces.Logger) logWrap {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return logWrap{
		logger: logger,
	}
}

func (l logWrap) Debug(msg string, args ...any) {
	l.logger.Debug(`[GoCAT] `+msg, args...)
}

func (l logWrap) Info(msg string, args ...any) {
	l.logger.Info(`[GoCAT] `+msg, args...)
}

func (l logWrap) Warn(msg string, args ...any) {
	l.logger.Warn(`[GoCAT] `+msg, args...)
}

func (l logWrap) Error(msg string, args ...any) {
	l.logger.Error(`[GoCAT] `+msg, args...)
}
