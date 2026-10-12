package dbtest

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"

	"github.com/pressly/goose/v3"
)

const migrationAttempts = 2

func migrate(ctx context.Context, run func(context.Context) ([]*goose.MigrationResult, error)) error {
	var err error
	for range migrationAttempts {
		_, err = run(ctx)
		if ctx.Err() != nil || !retryableMigration(err) {
			return err
		}
	}
	return err
}

func retryableMigration(err error) bool {
	return errors.Is(err, driver.ErrBadConn) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}
