package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/surkovvs/gocat/application/comp"
	"github.com/surkovvs/gocat/application/interfaces"
)

var (
	defaultProvidedSigs    = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	defaultShutdownTimeout = time.Second * 3
)

type (
	metadata struct {
		Name         string
		groupCounter int
	}
	sequentialGroup string
	maintain        struct {
		initTimeout     *time.Duration
		initRunCancel   context.CancelFunc
		initRunCtx      context.Context
		sigs            []os.Signal
		shutdownTimeout *time.Duration
		shutdownCtx     context.Context
		errFlow         chan error
		shutdownDone    chan struct{}
		exitCode        int
	}
	app struct {
		mtn     maintain
		modules map[comp.Component]sequentialGroup
		Meta    metadata
		logger  interfaces.Logger
	}
)

func New(opts ...appOption) *app {
	a := &app{
		Meta:    metadata{},
		logger:  nil,
		modules: make(map[comp.Component]sequentialGroup),
		mtn: maintain{
			initTimeout:   nil,
			initRunCancel: nil,
			initRunCtx:    nil,
			shutdownCtx:   context.Background(),
			errFlow:       make(chan error),
			shutdownDone:  make(chan struct{}),
		},
	}
	for _, opt := range opts {
		opt(a)
	}
	a.defaultSettingsCheckAndApply()
	go a.accompaniment()
	return a
}

func (a *app) defaultSettingsCheckAndApply() {
	if a.Meta.Name == "" {
		a.Meta.Name = `unnamed`
	}

	a.logger = newLogWrap(a.logger)

	if a.mtn.sigs == nil {
		a.mtn.sigs = defaultProvidedSigs
	}
	if a.mtn.shutdownTimeout == nil {
		a.mtn.shutdownTimeout = &defaultShutdownTimeout
	}
}

func (a *app) accompaniment() {
	syscallC := make(chan os.Signal, 1)
	signal.Notify(syscallC, a.mtn.sigs...)
	for {
		select {
		case err := <-a.mtn.errFlow:
			a.logger.Error(`module error`,
				"application", a.Meta.Name,
				`error`, err)
		case sig := <-syscallC:
			a.logger.Info(`graceful shutdown started by syscall`,
				"application", a.Meta.Name,
				`syscall`, sig.String())

			signal.Stop(syscallC)
			a.mtn.initRunCancel()
			go a.gracefulShutdown()
		}
	}
}
