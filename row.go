package pgy

import (
	"github.com/jackc/pgx/v5"
)

// rows

type rowsSt struct {
	pgx.Rows
}

func (o rowsSt) Err() error {
	return uHErr(o.Rows.Err())
}

func (o rowsSt) Scan(dest ...any) error {
	return uHErr(o.Rows.Scan(dest...))
}

// row

type rowSt struct {
	pgx.Row
}

func (o rowSt) Scan(dest ...any) error {
	return uHErr(o.Row.Scan(dest...))
}
