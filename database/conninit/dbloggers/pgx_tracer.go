package dbloggers

import (
	"context"

	pgx5 "github.com/jackc/pgx/v5"
	"github.com/surkovvs/gocat/database/conninit"
)

type PGX5Tracer struct {
	conninit.Logger
}

func NewPGX5Tracer(logger conninit.Logger) PGX5Tracer {
	return PGX5Tracer{logger}
}

// queries

func (tracer PGX5Tracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx5.Conn,
	data pgx5.TraceQueryStartData,
) context.Context {
	tracer.Logger.Debug("Query command executing",
		"sql", data.SQL,
		"args", data.Args)

	return ctx
}

func (tracer PGX5Tracer) TraceQueryEnd(_ context.Context, _ *pgx5.Conn, _ pgx5.TraceQueryEndData) {
}

// batch

func (tracer PGX5Tracer) TraceBatchStart(ctx context.Context, _ *pgx5.Conn, data pgx5.TraceBatchStartData) context.Context {
	tracer.Logger.Debug("Batch start",
		"batch len", data.Batch.Len(),
		"queries", data.Batch.QueuedQueries)
	return ctx
}

func (tracer PGX5Tracer) TraceBatchQuery(_ context.Context, _ *pgx5.Conn, data pgx5.TraceBatchQueryData) {
	tracer.Logger.Debug(
		"Batch query command executing",
		"sql", data.SQL,
		"args", data.Args,
		"error", data.Err)
}

func (tracer PGX5Tracer) TraceBatchEnd(ctx context.Context, conn *pgx5.Conn, data pgx5.TraceBatchEndData) {
}

// func (tracer defaultTracerPGX) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context

// func (tracer defaultTracerPGX) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData)

// func (tracer defaultTracerPGX) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context

// func (tracer defaultTracerPGX) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData)

// func (tracer defaultTracerPGX) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context

// func (tracer defaultTracerPGX) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData)
