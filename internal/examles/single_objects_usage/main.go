package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/surkovvs/ggwp"
	"github.com/surkovvs/ggwp/internal/examles/single_objects_usage/appsim"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ggwp.SetDefault(
		ggwp.NewShutdown(
			ggwp.WithLogger(slog.Default()),
			ggwp.WithStopContext(ctx),
			ggwp.WithShutdownTimeout(time.Second),
		),
	)

	appsim.Init()

	cancel()
	time.Sleep(time.Second * 3)
}
