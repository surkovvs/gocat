package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/surkovvs/ggwp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}))
	shutDowner := ggwp.NewShutdown(ggwp.WithLogger(logger), ggwp.WithShutdownTimeout(time.Second))

	shutDowner.RegisterNamedGracefulStop("first", func() error {
		time.Sleep(time.Second / 2)
		return nil
	})
	shutDowner.RegisterNamedGracefulStop("second", func() error {
		// time.Sleep(time.Second * 2)
		return nil
	})
	shutDowner.RegisterNamedGracefulStop("third", func() error {
		return errors.New("error from third")
	})

	shutDowner.RegisterGracefulStop(func() error {
		time.Sleep(time.Second / 2)
		return nil
	})
	shutDowner.RegisterGracefulStop(func() error {
		// time.Sleep(time.Second * 2)
		return nil
	})
	shutDowner.RegisterGracefulStop(func() error {
		return errors.New("error from unnamed")
	})

	time.Sleep(time.Second)
	shutDowner.GetStopTrigger()(1)

	time.Sleep(10 * time.Second)
	panic(`shutdown didnt worked`)
}
