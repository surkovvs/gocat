package pgx5connect

import (
	"context"
	"fmt"
	"time"

	"github.com/surkovvs/gocat/database/conninit"
	"github.com/surkovvs/gocat/database/conninit/dbloggers"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	MaxConnIdleTime       time.Duration
	MaxConns              int32
	MinConns              int32
	HealthCheckPeriod     time.Duration
}

func InitPGX5PoolCfg(ctx context.Context, cfg conninit.Config, pCfg *PoolConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port,
		cfg.DBName,
		cfg.User,
		cfg.Pass,
	)
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pool parse config: %w", err)
	}

	if cfg.Logger != nil {
		poolCfg.ConnConfig.Tracer = dbloggers.NewPGX5Tracer(cfg.Logger)
		poolCfg.BeforeConnect = func(_ context.Context, _ *pgx.ConnConfig) error {
			cfg.Logger.Debug("pgx pool: init new connection")
			return nil
		}
		poolCfg.AfterConnect = func(_ context.Context, c *pgx.Conn) error {
			cfg.Logger.Debug(
				"pgx pool: created new connection",
				"local address", c.PgConn().Conn().LocalAddr().String(),
				"backend PID", c.PgConn().PID(),
			)
			return nil
		}
		poolCfg.BeforeAcquire = func(_ context.Context, c *pgx.Conn) bool {
			cfg.Logger.Debug("pgx pool: acquired conn",
				"local address", c.PgConn().Conn().LocalAddr().String(),
				"backend PID", c.PgConn().PID(),
			)
			return true
		}
		poolCfg.AfterRelease = func(c *pgx.Conn) bool {
			cfg.Logger.Debug("pgx pool: released conn",
				"local address", c.PgConn().Conn().LocalAddr().String(),
				"backend PID", c.PgConn().PID(),
			)
			return true
		}
		poolCfg.BeforeClose = func(c *pgx.Conn) {
			cfg.Logger.Debug("pgx pool: closing conn",
				"local address", c.PgConn().Conn().LocalAddr().String(),
				"backend PID", c.PgConn().PID(),
			)
		}
	}

	poolCfg.ConnConfig.TLSConfig = cfg.TLS

	if pCfg != nil {
		poolCfg.MaxConnLifetime = pCfg.MaxConnLifetime
		poolCfg.MaxConnLifetimeJitter = pCfg.MaxConnLifetimeJitter
		poolCfg.MaxConnIdleTime = pCfg.MaxConnIdleTime
		poolCfg.MaxConns = pCfg.MaxConns
		poolCfg.MinConns = pCfg.MinConns
		poolCfg.HealthCheckPeriod = pCfg.HealthCheckPeriod
	}

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
