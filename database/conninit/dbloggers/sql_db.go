package dbloggers

import (
	"context"

	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/surkovvs/gocat/database/conninit"
)

type SQLDBLogger struct {
	conninit.Logger
}

func NewSQLDBLogger(logger conninit.Logger) sqldblogger.Logger {
	return SQLDBLogger{logger}
}

func (l SQLDBLogger) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
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

// LogLevelTrace = 6
// LogLevelDebug = 5
// LogLevelInfo  = 4
// LogLevelWarn  = 3
// LogLevelError = 2
// LogLevelNone  = 1
