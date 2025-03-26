package catdefpgx

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/surkovvs/gocat/catcfg"
	"github.com/surkovvs/gocat/catlog"
)

// TODO: implement methods
func InitPGXPool(cfg catcfg.Config) (*pgx.ConnPool, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port,
		cfg.Name,
		cfg.User,
		cfg.Pass,
	)
	connCfg, err := pgx.ParseConnectionString(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.Logger != nil {
		connCfg.LogLevel = 6
		connCfg.Logger = logAdapter{cfg.Logger}
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connCfg,
	})
	if err != nil {
		return nil, err
	}
	return pool, nil
}

type logAdapter struct {
	catlog.Logger
}

func (l logAdapter) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
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
