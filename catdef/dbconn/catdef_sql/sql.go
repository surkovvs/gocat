package catdefsql

import (
	"context"
	"database/sql"

	"github.com/surkovvs/gocat/catcfg"
	"github.com/surkovvs/gocat/catlog"

	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type Database struct {
	*sql.DB
}

func NewPQDatabase(cfg catcfg.Config) (*Database, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	if cfg.Logger != nil {
		db = sqldblogger.OpenDriver(cfg.GetDSN(), db.Driver(), logAdapter{cfg.Logger})
	}

	return &Database{db}, nil
}

func (db *Database) Init(ctx context.Context) error {
	return db.PingContext(ctx)
}

func (db *Database) Shutdown(_ context.Context) error {
	return db.Close()
}

type logAdapter struct {
	catlog.Logger
}

func (l logAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
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
