package interfaces

import "context"

type (
	Healthchecker interface {
		Healthcheck(ctx context.Context) error
	}
	Initializer interface {
		Init(ctx context.Context) error
	}
	Runner interface {
		Run(ctx context.Context) error
	}
	Shutdowner interface {
		Shutdown(ctx context.Context) error
	}
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
