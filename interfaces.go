package pgy

import (
	"context"
)

type ConnectionI interface {
	Exec(ctx context.Context, sql string, args ...any) error
	Query(ctx context.Context, sql string, args ...any) (RowsI, error)
	QueryRow(ctx context.Context, sql string, args ...any) RowI
	ExecM(ctx context.Context, sql string, argMap map[string]any) error
	QueryM(ctx context.Context, sql string, argMap map[string]any) (RowsI, error)
	QueryRowM(ctx context.Context, sql string, argMap map[string]any) RowI
	GetTransactionManager() TransactionManagerI
}

type TransactionManagerI interface {
	TxFn(ctx context.Context, f func(context.Context) error) error
	RenewContextTx(ctx context.Context) error
	TxAddAsyncCallback(ctx context.Context, f func())
}

type RowsI interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type RowI interface {
	Scan(dest ...any) error
}
