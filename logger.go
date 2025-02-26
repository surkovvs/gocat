package shutdowner

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type nullLog struct{}

func (nullLog) Info(msg string, args ...any) {
}

func (nullLog) Error(msg string, args ...any) {
}
