package pgy

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // driver
)

type Connection struct {
	Con                *pgxpool.Pool
	TransactionManager *TransactionManager
}

func NewConnection(dsn string) (*Connection, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("fail to parse dsn: %w", err)
	}

	//cfg.ConnConfig.RuntimeParams["timezone"] = opts.Timezone

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("fail to create pool: %w", err)
	}

	connection := &Connection{
		Con: pool,
	}

	connection.TransactionManager = NewTransactionManager(connection.Con)

	return connection, nil
}

func (c *Connection) Exec(ctx context.Context, sql string, args ...any) error {
	var err error

	if tx := c.TransactionManager.extractTxFromContext(ctx); tx != nil {
		_, err = tx.tx.Exec(ctx, sql, args...)
	} else {
		_, err = c.Con.Exec(ctx, sql, args...)
	}

	return uHErr(err)
}

func (c *Connection) Query(ctx context.Context, sql string, args ...any) (RowsI, error) {
	var err error
	var rows pgx.Rows

	if tx := c.TransactionManager.extractTxFromContext(ctx); tx != nil {
		rows, err = tx.tx.Query(ctx, sql, args...)
	} else {
		rows, err = c.Con.Query(ctx, sql, args...)
	}

	return rowsSt{Rows: rows}, uHErr(err)
}

func (c *Connection) QueryRow(ctx context.Context, sql string, args ...any) RowI {
	var row pgx.Row

	if tx := c.TransactionManager.extractTxFromContext(ctx); tx != nil {
		row = tx.tx.QueryRow(ctx, sql, args...)
	} else {
		row = c.Con.QueryRow(ctx, sql, args...)
	}

	return rowSt{Row: row}
}

func (c *Connection) ExecM(ctx context.Context, sql string, argMap map[string]any) error {
	rbSql, args := uQueryRebindNamed(sql, argMap)
	return c.Exec(ctx, rbSql, args...)
}

func (c *Connection) QueryM(ctx context.Context, sql string, argMap map[string]any) (RowsI, error) {
	rbSql, args := uQueryRebindNamed(sql, argMap)
	return c.Query(ctx, rbSql, args...)
}

func (c *Connection) QueryRowM(ctx context.Context, sql string, argMap map[string]any) RowI {
	rbSql, args := uQueryRebindNamed(sql, argMap)
	return c.QueryRow(ctx, rbSql, args...)
}

func (c *Connection) GetTransactionManager() TransactionManagerI {
	return c.TransactionManager
}
