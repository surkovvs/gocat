package ggwp

import (
	"context"
	"os"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var (
	defaultSDSigs                = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	defaultTimeout time.Duration = time.Second * 3
)

type nullLog struct{}

func (nullLog) Info(msg string, args ...any) {
}

func (nullLog) Error(msg string, args ...any) {
}

type zapWrap struct {
	z *zap.SugaredLogger
}

func (w zapWrap) Info(msg string, args ...any) {
	w.z.Infow(msg, args...)
}

func (w zapWrap) Error(msg string, args ...any) {
	w.z.Errorw(msg, args...)
}

func WithLogger(logger Logger) shutdownOpt {
	return func(s *shutdown) {
		s.logger = logger
	}
}

func WithZapLogger(logger *zap.SugaredLogger) shutdownOpt {
	return func(s *shutdown) {
		s.logger = zapWrap{logger}
	}
}

func WithShutdownTimeout(dur time.Duration) shutdownOpt {
	return func(s *shutdown) {
		s.timeout = dur
	}
}

func WithStopContext(stopCtx context.Context) shutdownOpt {
	return func(s *shutdown) {
		s.stopCtx = stopCtx
	}
}

func WithProvidedSigs(sigs ...os.Signal) shutdownOpt {
	return func(s *shutdown) {
		s.sigs = sigs
	}
}

func (sd *shutdown) defaultSettingsCheckAndApply() {
	if sd.logger == nil {
		sd.logger = nullLog{}
	}

	if sd.timeout == 0 {
		sd.timeout = defaultTimeout
	}

	if len(sd.sigs) == 0 {
		sd.sigs = defaultSDSigs
	}

	if sd.stopCtx == nil {
		sd.stopCtx = context.Background()
	}
}
