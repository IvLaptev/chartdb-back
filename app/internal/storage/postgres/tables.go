package postgres

import (
	"context"
)

var Tables []string = []string{
	"diagrams",
	"users",
	"user_confirmations",
}

func (s *Storage) Erase(ctx context.Context) {
	for _, table := range Tables {
		s.db.ExecContext(ctx, "TRUNCATE TABLE "+table+" CASCADE")
	}
}
