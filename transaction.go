package pgy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // driver
)

type txCtxKeyT int8

const (
	txCtxKey = txCtxKeyT(1)
)

type Tx struct {
	tx             pgx.Tx
	asyncCallbacks []func()
}

type TransactionManager struct {
	connection *pgxpool.Pool
}

func NewTransactionManager(connection *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		connection: connection,
	}
}

func (t *TransactionManager) extractTxFromContext(ctx context.Context) *Tx {
	contextV := ctx.Value(txCtxKey)
	if contextV == nil {
		return nil
	}

	if tx, ok := contextV.(*Tx); ok {
		return tx
	}

	return nil
}

func (t *TransactionManager) contextWithTx(ctx context.Context) (context.Context, error) {
	if t.extractTxFromContext(ctx) != nil {
		return ctx, nil
	}

	tx, err := t.connection.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("fail to begin transaction: %w", err)
	}

	return context.WithValue(ctx, txCtxKey, &Tx{tx: tx}), nil
}

func (t *TransactionManager) commitContextTx(ctx context.Context) error {
	tx := t.extractTxFromContext(ctx)
	if tx == nil {
		return nil
	}

	err := tx.tx.Commit(ctx)
	if err != nil {
		_ = tx.tx.Rollback(ctx)

		return fmt.Errorf("fail to commit transaction: %w", err)
	}

	// run async callbacks
	go t.callbackRunner(tx.asyncCallbacks)

	return nil
}

func (t *TransactionManager) rollbackContextTx(ctx context.Context) {
	tx := t.extractTxFromContext(ctx)
	if tx == nil {
		return
	}

	_ = tx.tx.Rollback(ctx)
}

func (t *TransactionManager) RenewContextTx(ctx context.Context) error {
	var err error

	tx := t.extractTxFromContext(ctx)
	if tx != nil {
		err = tx.tx.Commit(ctx)
		if err != nil {
			_ = tx.tx.Rollback(ctx)

			return fmt.Errorf("fail to commit transaction: %w", err)
		}
	}

	tx.tx, err = t.connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("fail to begin transaction: %w", err)
	}

	return nil
}

func (t *TransactionManager) TxFn(ctx context.Context, f func(context.Context) error) error {
	var err error

	if ctx == nil {
		ctx = context.Background()
	}

	if ctx, err = t.contextWithTx(ctx); err != nil {
		return err
	}
	defer func() { t.rollbackContextTx(ctx) }()

	err = f(ctx)
	if err != nil {
		return err
	}

	return t.commitContextTx(ctx)
}

func (t *TransactionManager) TxAddAsyncCallback(ctx context.Context, f func()) {
	tx := t.extractTxFromContext(ctx)
	if tx == nil {
		go t.callbackRunner([]func(){f})
	} else {
		tx.asyncCallbacks = append(tx.asyncCallbacks, f)
	}
}

func (t *TransactionManager) callbackRunner(callbacks []func()) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Tx.callbackRunner recovered from panic", "error", err)
		}
	}()

	for _, f := range callbacks {
		f()
	}
}
