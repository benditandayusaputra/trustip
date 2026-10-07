package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	maxConnections          = 10
	maxConnectionIdleTime   = 3 * time.Minute
	readTimeout             = 3 * time.Second
	writeTimeout            = 10 * time.Second
	unavailableRetrySeconds = 5
	slowQueryThreshold      = 200 * time.Millisecond
	uniqueViolationCode     = "23505"
)

var (
	ErrInvalidURL  = errors.New("database url cannot be parsed")
	ErrSchemaOwner = errors.New("server must not connect as the schema owner")
)

type Clock interface {
	Now() time.Time
}

type Options struct {
	URL               string
	Logger            *slog.Logger
	Clock             Clock
	RefuseSchemaOwner bool
}

type Pool struct {
	pool *pgxpool.Pool
}

type txKey struct{}

func Open(ctx context.Context, opts Options) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}
	cfg.MaxConns = maxConnections
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = maxConnectionIdleTime
	cfg.ConnConfig.Tracer = newSlowQueryTracer(opts.Logger, opts.Clock, slowQueryThreshold)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	opened := &Pool{pool: pool}
	if !opts.RefuseSchemaOwner {
		return opened, nil
	}
	if err := opened.refuseSchemaOwner(ctx); err != nil {
		opened.Close()
		return nil, err
	}
	return opened, nil
}

func (p *Pool) refuseSchemaOwner(ctx context.Context) error {
	readCtx, cancel := ReadContext(ctx)
	defer cancel()
	isOwner, err := sqlc.New(p.pool).IsSchemaOwner(readCtx)
	if err != nil {
		return fmt.Errorf("check database role: %w", unavailable(err))
	}
	if isOwner {
		return ErrSchemaOwner
	}
	return nil
}

func (p *Pool) Close() {
	p.pool.Close()
}

func (p *Pool) Ping(ctx context.Context) error {
	readCtx, cancel := ReadContext(ctx)
	defer cancel()
	if err := p.pool.Ping(readCtx); err != nil {
		return unavailable(err)
	}
	return nil
}

func (p *Pool) Querier(ctx context.Context) sqlc.DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return p.pool
}

func (p *Pool) WithinTx(ctx context.Context, work func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return work(ctx)
	}
	writeCtx, cancel := WriteContext(ctx)
	defer cancel()
	tx, err := p.pool.Begin(writeCtx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", unavailable(err))
	}
	if err := work(context.WithValue(writeCtx, txKey{}, tx)); err != nil {
		return errors.Join(err, rollback(writeCtx, tx))
	}
	if err := tx.Commit(writeCtx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func rollback(ctx context.Context, tx pgx.Tx) error {
	err := tx.Rollback(context.WithoutCancel(ctx))
	if err == nil || errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return fmt.Errorf("rollback transaction: %w", err)
}

func ReadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, readTimeout)
}

func WriteContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, writeTimeout)
}

func unavailable(cause error) error {
	return apperror.Unavailable(unavailableRetrySeconds).WithCause(cause)
}

func IsUniqueViolation(err error, constraint string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == constraint
}
