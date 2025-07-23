package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const userConfirmationTable = "user_confirmations"

var (
	userConfirmationFields = []string{fieldID, fieldUserID, fieldCreatedAt, fieldExpiresAt}

	returningUserConfirmation = returning + strings.Join(userConfirmationFields, separator)
)

type userConfirmationEntity struct {
	ID        model.UserConfirmationID `db:"id"`
	UserID    model.UserID             `db:"user_id"`
	CreatedAt time.Time                `db:"created_at"`
	ExpiresAt time.Time                `db:"expires_at"`
}

func (s *Storage) GetUserConfirmationByID(ctx context.Context, id model.UserConfirmationID) (*model.UserConfirmation, error) {
	query := sq.Select(userConfirmationFields...).
		From(userConfirmationTable).
		Where(sq.Eq{fieldID: id.String()}).
		PlaceholderFormat(sq.Dollar)

	sql, args := query.MustSql()

	var entity userConfirmationEntity
	if err := sqlx.GetContext(ctx, s.DB(ctx), &entity, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return userConfirmationEntityToModel(entity), nil
}

func (s *Storage) CreateUserConfirmation(ctx context.Context, params *storage.CreateUserConfirmationParams) (*model.UserConfirmation, error) {
	now := time.Now()
	query := sq.Insert(userConfirmationTable).
		Columns(userConfirmationFields...).
		Values(
			params.ID,
			params.UserID,
			now,
			now.Add(params.Duration),
		).
		Suffix(returningUserConfirmation).
		PlaceholderFormat(sq.Dollar)

	sql, args := query.MustSql()
	var entity userConfirmationEntity
	if err := sqlx.GetContext(ctx, s.DB(ctx), &entity, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return userConfirmationEntityToModel(entity), nil
}

func userConfirmationEntityToModel(entity userConfirmationEntity) *model.UserConfirmation {
	return &model.UserConfirmation{
		ID:        entity.ID,
		UserID:    entity.UserID,
		CreatedAt: entity.CreatedAt,
		ExpiresAt: entity.ExpiresAt,
	}
}
