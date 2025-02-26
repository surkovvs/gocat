package ggwp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

const unnamed = "unnamed"

type stop struct {
	name     string
	stopFunc func() error
	finished bool
}

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type shutdown struct {
	stops    []*stop
	sigs     []os.Signal
	logger   Logger
	stopCtx  context.Context
	stopC    chan struct{}
	finish   chan struct{}
	timeout  time.Duration
	exitCode int
}

type shutdownOpt func(*shutdown)

type namedError struct {
	name string
	err  error
}

func NewShutdown(opts ...shutdownOpt) *shutdown {
	sd := &shutdown{
		stops:    make([]*stop, 0),
		sigs:     make([]os.Signal, 0),
		finish:   make(chan struct{}),
		stopC:    make(chan struct{}),
		exitCode: 0,
	}

	for _, opt := range opts {
		opt(sd)
	}
	sd.defaultSettingsCheckAndApply()
	sd.init()
	return sd
}

func (sd *shutdown) init() {
	syscallC := make(chan os.Signal, 1)
	signal.Notify(syscallC, sd.sigs...)

	go func() {
		select {
		case <-sd.stopCtx.Done():
			sd.logger.Info("graceful shutdown started by stop context")
		case sig := <-syscallC:
			sd.logger.Info("graceful shutdown started by syscall", "syscall", sig.String())
		case <-sd.stopC:
			sd.logger.Info("graceful shutdown started by stop function")
		}
		sd.gracefulShutdown()
	}()
}

func (sd *shutdown) RegisterGracefulStop(f func() error) {
	sd.stops = append(sd.stops, &stop{
		name:     unnamed,
		stopFunc: f,
	})
}

func (sd *shutdown) RegisterNamedGracefulStop(name string, f func() error) {
	sd.stops = append(sd.stops, &stop{
		name:     name,
		stopFunc: f,
	})
}

func (sd *shutdown) GetStopTrigger() func(exitCode int) {
	return func(exitCode int) {
		sd.exitCode = exitCode
		close(sd.stopC)
		<-sd.finish
	}
}

func (sd *shutdown) gracefulShutdown() {
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errChan := make(chan namedError, 1)
	ctx, cancel := context.WithTimeout(context.Background(), sd.timeout)
	defer cancel()
	for _, s := range sd.stops {
		wg.Add(1)
		go func(s *stop) {
			if err := s.stopFunc(); err != nil {
				errChan <- namedError{
					name: s.name,
					err:  err,
				}
			}
			s.finished = true
			wg.Done()
		}(s)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	go func() {
		for namedErr := range errChan {
			if namedErr.err != nil {
				errKey := `error from unnamed stop func`
				if namedErr.name != unnamed {
					errKey = fmt.Sprintf(`error from stop func "%s"`, namedErr.name)
				}
				sd.logger.Error("graceful stop", errKey, namedErr.err)
			}
		}

		close(done)
	}()

	select {
	case <-done:
		sd.logger.Info(`gracedul shutdown finished for all registered functions`, `functions count`, len(sd.stops))
	case <-ctx.Done():
		namesOfUnfinished := make([]string, 0, len(sd.stops))
		for _, stop := range sd.stops {
			if !stop.finished {
				namesOfUnfinished = append(namesOfUnfinished, stop.name)
			}
		}
		sd.logger.Error(`gracedul shutdown timeout exeeded`, `unfinished functions`, namesOfUnfinished)
	}
	os.Exit(sd.exitCode)
}
