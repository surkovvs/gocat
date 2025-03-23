package dbloggers

import (
	"github.com/surkovvs/gocat/database/conninit"

	"github.com/jackc/pgx"
)

type PGXLogger struct {
	conninit.Logger
}

func NewPGXLogger(logger conninit.Logger) pgx.Logger {
	return PGXLogger{logger}
}

func (l PGXLogger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]any, 0, len(data))
	for k, v := range data {
		fields = append(fields, k, v)
	}
	switch level {
	case 6:
		l.Logger.Debug(msg, fields...)
	case 5:
		l.Logger.Debug(msg, fields...)
	case 4:
		l.Logger.Info(msg, fields...)
	case 3:
		l.Logger.Warn(msg, fields...)
	case 2:
		l.Logger.Error(msg, fields...)
	}
}
