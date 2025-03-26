package catdefpgxp

import (
	"context"
	"fmt"

	"github.com/surkovvs/gocat/catcfg"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	cfg *pgxpool.Config
	*pgxpool.Pool
}

func (pool *Pool) New(cfg catcfg.Config) (*Pool, error) {
	cfgPool, err := pgxpool.ParseConfig(cfg.GetDSN())
	if cfg.Logger != nil {
		return nil, fmt.Errorf("pool parse config: %w", err)
	}

	if cfg.Logger != nil {
		if cfg.ConfigDB.LogQueries {
			cfgPool.ConnConfig.Tracer = NewPGX5Tracer(cfg.Logger)
		}
		if cfg.ConfigDB.ConfigPool.LogConnectOperations {
			cfgPool.BeforeConnect = func(_ context.Context, _ *pgx.ConnConfig) error {
				cfg.Logger.Debug("pgx pool: init new connection")
				return nil
			}
			cfgPool.AfterConnect = func(_ context.Context, c *pgx.Conn) error {
				cfg.Logger.Debug(
					"pgx pool: created new connection",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return nil
			}
			cfgPool.BeforeAcquire = func(_ context.Context, c *pgx.Conn) bool {
				cfg.Logger.Debug("pgx pool: acquired conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return true
			}
			cfgPool.AfterRelease = func(c *pgx.Conn) bool {
				cfg.Logger.Debug("pgx pool: released conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return true
			}
			cfgPool.BeforeClose = func(c *pgx.Conn) {
				cfg.Logger.Debug("pgx pool: closing conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
			}
		}
	}

	SetIfNotNil(&cfgPool.MaxConnLifetime, cfg.MaxConnLifetime)
	SetIfNotNil(&cfgPool.MaxConnLifetimeJitter, cfg.MaxConnLifetimeJitter)
	SetIfNotNil(&cfgPool.MaxConnIdleTime, cfg.MaxConnIdleTime)
	SetIfNotNil(&cfgPool.MaxConns, cfg.MaxConns)
	SetIfNotNil(&cfgPool.MinConns, cfg.MinConns)
	SetIfNotNil(&cfgPool.HealthCheckPeriod, cfg.HealthCheckPeriod)

	return &Pool{
		cfg: cfgPool,
	}, nil
}

func SetIfNotNil[T any](a *T, b *T) {
	if b != nil {
		*a = *b
	}
}

func (pool *Pool) Init(ctx context.Context) error {
	var err error
	pool.Pool, err = pgxpool.NewWithConfig(ctx, pool.cfg)
	return err
}

func (pool *Pool) Shutdown(_ context.Context) error {
	pool.Pool.Close()
	return nil
}
