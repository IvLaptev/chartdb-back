package postgres

import (
	"database/sql"
	"errors"

	"github.com/IvLaptev/chartdb-back/internal/storage"
)

func formatError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrNotFound
	}

	return err
}
