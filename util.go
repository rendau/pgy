package pgy

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func uHErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, sql.ErrNoRows):
		err = sql.ErrNoRows
	default:
		err = fmt.Errorf("pg-error: %w", err)
	}

	return err
}

func uQueryRebindNamed(sql string, argMap map[string]any) (string, []any) {
	args := make([]any, 0, len(argMap))

	for k, v := range argMap {
		if strings.Contains(sql, "${"+k+"}") {
			args = append(args, v)
			sql = strings.ReplaceAll(sql, "${"+k+"}", "$"+strconv.Itoa(len(args)))
		}
	}

	return sql, args
}
