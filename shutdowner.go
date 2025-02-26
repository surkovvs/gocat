package shutdowner

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

func NewShutdown(opts ...ShutDownerOpt) *shutdown {
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
			sd.logger.Info("graceful shutdown started by syscall", sig.String())
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

func (sd *shutdown) GetStopFunc() func(exitCode int) {
	return func(exitCode int) {
		sd.exitCode = exitCode
		close(sd.stopC)
		<-sd.finish
	}
}

func (sd *shutdown) gracefulShutdown() {
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errChan := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), sd.timeout)
	defer cancel()
	for _, s := range sd.stops {
		wg.Add(1)
		go func(s *stop) {
			if err := s.stopFunc(); err != nil {
				if s.name != unnamed {
					errChan <- fmt.Errorf(`for registered function "%s": %w`, s.name, err)
				} else {
					errChan <- err
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
		for err := range errChan {
			if err != nil {
				sd.logger.Error("graceful stop error", err)
			}
		}

		close(done)
	}()

	select {
	case <-done:
		sd.logger.Info("gracedul shutdown finished for all registered functions")
	case <-ctx.Done():
		namesOfUnfinished := make([]string, 0, len(sd.stops))
		for _, stop := range sd.stops {
			if !stop.finished {
				namesOfUnfinished = append(namesOfUnfinished, stop.name)
			}
		}
		sd.logger.Error("gracedul shutdown timeout exeeded, unfinished functions:", namesOfUnfinished)
	}
	os.Exit(sd.exitCode)
}
