package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient interface {
	Close()
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type PostgresClientImp struct {
	pool *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, dsn string) (PostgresClient, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("erro criar pool: %w", err)
	}
	return &PostgresClientImp{pool: pool}, nil
}

func (p *PostgresClientImp) Close() {
	p.pool.Close()
}

func (p *PostgresClientImp) Exec(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	cmdTag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("erro executar comando: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (p *PostgresClientImp) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *PostgresClientImp) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}
