package main

import (
	"errors"
	"os"
	"time"

	"github.com/surkovvs/ggwp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), os.Stdout, zap.InfoLevel))
	shutDowner := ggwp.NewShutdown(ggwp.WithZapLogger(logger.Sugar()), ggwp.WithShutdownTimeout(time.Second))

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
