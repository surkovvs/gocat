package pgxconnect

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/surkovvs/gocat/database/conninit"
	"github.com/surkovvs/gocat/database/conninit/dbloggers"
)

func InitPGXPool(cfg conninit.Config) (*pgx.ConnPool, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port,
		cfg.DBName,
		cfg.User,
		cfg.Pass,
	)
	connCfg, err := pgx.ParseConnectionString(dsn)
	if err != nil {
		return nil, err
	}

	connCfg.TLSConfig = cfg.TLS

	if cfg.Logger != nil {
		connCfg.LogLevel = 6
		connCfg.Logger = dbloggers.NewPGXLogger(cfg.Logger)
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connCfg,
	})
	if err != nil {
		return nil, err
	}
	return pool, nil
}
