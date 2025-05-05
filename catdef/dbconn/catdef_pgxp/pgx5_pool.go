package catdefpgxp

import (
	"context"
	"fmt"

	"github.com/surkovvs/gocat/catcfg"
	"github.com/surkovvs/gocat/catlog"

	pgx5 "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Pool struct {
		cfg *pgxpool.Config
		*pgxpool.Pool
	}

	tracer struct {
		catlog.Logger
	}
)

func New(cfg catcfg.Config) (*Pool, error) {
	cfgPool, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("pool parse config: %w", err)
	}

	if cfg.Logger != nil {
		if cfg.ConfigDB.LogQueries {
			cfgPool.ConnConfig.Tracer = tracer{cfg.Logger}
		}
		if cfg.ConfigDB.ConfigPool.LogConnectOperations {
			cfgPool.BeforeConnect = func(_ context.Context, _ *pgx5.ConnConfig) error {
				cfg.Logger.Debug("pgx pool: init new connection")
				return nil
			}
			cfgPool.AfterConnect = func(_ context.Context, c *pgx5.Conn) error {
				cfg.Logger.Debug(
					"pgx pool: created new connection",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return nil
			}
			cfgPool.BeforeAcquire = func(_ context.Context, c *pgx5.Conn) bool {
				cfg.Logger.Debug("pgx pool: acquired conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return true
			}
			cfgPool.AfterRelease = func(c *pgx5.Conn) bool {
				cfg.Logger.Debug("pgx pool: released conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
				return true
			}
			cfgPool.BeforeClose = func(c *pgx5.Conn) {
				cfg.Logger.Debug("pgx pool: closing conn",
					"local address", c.PgConn().Conn().LocalAddr().String(),
					"backend PID", c.PgConn().PID(),
				)
			}
		}
	}

	setIfNotNil(&cfgPool.MaxConnLifetime, cfg.MaxConnLifetime)
	setIfNotNil(&cfgPool.MaxConnLifetimeJitter, cfg.MaxConnLifetimeJitter)
	setIfNotNil(&cfgPool.MaxConnIdleTime, cfg.MaxConnIdleTime)
	setIfNotNil(&cfgPool.MaxConns, cfg.MaxConns)
	setIfNotNil(&cfgPool.MinConns, cfg.MinConns)
	setIfNotNil(&cfgPool.HealthCheckPeriod, cfg.HealthCheckPeriod)

	return &Pool{
		cfg: cfgPool,
	}, nil
}

func setIfNotNil[T any](a *T, b *T) {
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

// queries

func (tracer tracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx5.Conn,
	data pgx5.TraceQueryStartData,
) context.Context {
	tracer.Logger.Debug("Query command executing",
		"sql", data.SQL,
		"args", data.Args)

	return ctx
}

func (tracer tracer) TraceQueryEnd(_ context.Context, _ *pgx5.Conn, _ pgx5.TraceQueryEndData) {
}

// batch

func (tracer tracer) TraceBatchStart(ctx context.Context, _ *pgx5.Conn, data pgx5.TraceBatchStartData) context.Context {
	tracer.Logger.Debug("Batch start",
		"batch len", data.Batch.Len(),
		"queries", data.Batch.QueuedQueries)
	return ctx
}

func (tracer tracer) TraceBatchQuery(_ context.Context, _ *pgx5.Conn, data pgx5.TraceBatchQueryData) {
	tracer.Logger.Debug(
		"Batch query command executing",
		"sql", data.SQL,
		"args", data.Args,
		"error", data.Err)
}

func (tracer tracer) TraceBatchEnd(ctx context.Context, conn *pgx5.Conn, data pgx5.TraceBatchEndData) {
}

// func (tracer defaultTracerPGX) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context

// func (tracer defaultTracerPGX) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData)

// func (tracer defaultTracerPGX) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context

// func (tracer defaultTracerPGX) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData)

// func (tracer defaultTracerPGX) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context

// func (tracer defaultTracerPGX) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData)
