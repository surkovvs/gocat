package shutdowner

import (
	"context"
	"os"
	"syscall"
	"time"
)

var (
	defaultSDSigs                = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	defaultTimeout time.Duration = time.Second * 3
)

type ShutDownerOpt func(*shutdown)

func WithLogger(logger Logger) ShutDownerOpt {
	return func(s *shutdown) {
		s.logger = logger
	}
}

func WithShutdownTimeout(dur time.Duration) ShutDownerOpt {
	return func(s *shutdown) {
		s.timeout = dur
	}
}

func WithStopContext(stopCtx context.Context) ShutDownerOpt {
	return func(s *shutdown) {
		s.stopCtx = stopCtx
	}
}

func StopOnSigs(sigs ...os.Signal) ShutDownerOpt {
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
