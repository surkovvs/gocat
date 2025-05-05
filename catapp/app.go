package catapp

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/surkovvs/gocat/catapp/compstor"
	"github.com/surkovvs/gocat/catapp/interfaces"
)

var (
	defaultProvidedSigs    = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	defaultShutdownTimeout = time.Second * 3
	privelegedGroupName    = `global`
)

type (
	execution struct {
		done          chan struct{}
		errFlow       chan error
		initRunCancel context.CancelFunc
		initTimeout   *time.Duration
	}
	shutdown struct {
		ctx          context.Context
		shutdownDone chan struct{}
		sigs         []os.Signal
		timeout      *time.Duration
		exitCode     int
	}
	app struct {
		execution execution
		shutdown  shutdown
		storage   compstor.CompsStorage
		name      string
		logger    interfaces.Logger
	}
)

func New(opts ...appOption) *app {
	a := &app{
		execution: execution{
			done:          make(chan struct{}),
			errFlow:       make(chan error),
			initRunCancel: nil,
			initTimeout:   nil,
		},
		shutdown: shutdown{
			ctx:          context.Background(),
			shutdownDone: make(chan struct{}),
			sigs:         nil,
			timeout:      nil,
			exitCode:     0,
		},
		storage: compstor.NewCompsStorage(),
		name:    "",
		logger:  nil,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.defaultSettingsCheckAndApply()
	go a.accompaniment()

	if err := a.storage.AddGroup(privelegedGroupName); err != nil && !errors.Is(err, compstor.ErrGroupAlreadyRegistered) {
		log.Fatal(err)
	}

	return a
}

func (a *app) defaultSettingsCheckAndApply() {
	if a.name == "" {
		a.name = `unnamed`
	}

	a.logger = newLogWrap(a.logger)

	if a.shutdown.sigs == nil {
		a.shutdown.sigs = defaultProvidedSigs
	}
	if a.shutdown.timeout == nil {
		a.shutdown.timeout = &defaultShutdownTimeout
	}
}

func (a *app) accompaniment() {
	syscallC := make(chan os.Signal, 1)
	signal.Notify(syscallC, a.shutdown.sigs...)
	for {
		select {
		case err := <-a.execution.errFlow:
			a.logger.Error(`module error`,
				"application", a.name,
				`error`, err)
		case <-a.execution.done:
			a.logger.Debug(`execution finished graceful shutdown started`,
				"application", a.name)
			signal.Stop(syscallC)
			a.execution.initRunCancel()
			go a.gracefulShutdown()
		case sig := <-syscallC:
			a.logger.Info(`graceful shutdown started by syscall`,
				"application", a.name,
				`syscall`, sig.String())

			signal.Stop(syscallC)
			a.execution.initRunCancel()
			go a.gracefulShutdown()
		}
	}
}
