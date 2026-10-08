package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const checkViolationCode = "23514"

func IsCheckViolation(err error, constraint string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == checkViolationCode && pgErr.ConstraintName == constraint
}
