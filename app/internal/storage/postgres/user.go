package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const userTable = "users"

var (
	userColumns = []string{fieldID, fieldLogin, fieldPasswordHash, fieldType,
		fieldConfirmedAt, fieldCreatedAt, fieldUpdatedAt, fieldDeletedAt}

	returningUser = returning + strings.Join(userColumns, separator)
)

type userEntity struct {
	ID           model.UserID `db:"id"`
	Login        string       `db:"login"`
	PasswordHash *string      `db:"password_hash"`
	Type         string       `db:"type"`
	ConfirmedAt  *time.Time   `db:"confirmed_at"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
	DeletedAt    *time.Time   `db:"deleted_at"`
}

func (s *Storage) GetUserByID(ctx context.Context, id model.UserID) (*model.User, error) {
	query := sq.Select(userColumns...).
		From(userTable).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: id.String()}).
		PlaceholderFormat(sq.Dollar)

	sql, args := query.MustSql()

	var user userEntity
	if err := sqlx.GetContext(ctx, s.DB(ctx), &user, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return userEntityToModel(&user)
}

func (s *Storage) GetAllUsers(ctx context.Context, filter []*model.FilterTerm) ([]*model.User, error) {
	query := sq.Select(userColumns...).
		From(userTable).
		Where(sq.Eq{fieldDeletedAt: nil}).
		PlaceholderFormat(sq.Dollar)

	query, err := filterQuery(query, userTable, filter)
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}

	sql, args := query.MustSql()

	var users []*userEntity
	if err := sqlx.SelectContext(ctx, s.DB(ctx), &users, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return makeUserList(users)
}

func (s *Storage) CreateUser(ctx context.Context, params *storage.CreateUserParams) (*model.User, error) {
	now := time.Now()

	query := sq.Insert(userTable).
		Columns(userColumns...).
		Values(
			params.ID,
			params.Login,
			params.PasswordHash,
			params.Type.String(),
			nil,
			now,
			now,
			nil,
		).
		Suffix(returningUser).
		PlaceholderFormat(sq.Dollar)

	sql, args := query.MustSql()

	var user userEntity
	if err := sqlx.GetContext(ctx, s.DB(ctx), &user, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return userEntityToModel(&user)
}

func (s *Storage) PatchUser(ctx context.Context, params *storage.PatchUserParams) (*model.User, error) {
	now := time.Now()

	query := sq.Update(userTable).
		Set(fieldUpdatedAt, now).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: params.ID.String()}).
		Suffix(returningUser).
		PlaceholderFormat(sq.Dollar)

	query = patchQueryOptional(query, fieldConfirmedAt, params.ConfirmedAt)

	sql, args := query.MustSql()

	var user userEntity
	if err := sqlx.GetContext(ctx, s.DB(ctx), &user, sql, args...); err != nil {
		return nil, formatError(err)
	}

	return userEntityToModel(&user)
}

func userEntityToModel(user *userEntity) (*model.User, error) {
	userType, err := model.UserTypeFromString(user.Type)
	if err != nil {
		return nil, fmt.Errorf("user type from string: %w", err)
	}

	return &model.User{
		ID:           user.ID,
		Login:        user.Login,
		PasswordHash: user.PasswordHash,
		Type:         userType,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func makeUserList(users []*userEntity) ([]*model.User, error) {
	result := make([]*model.User, 0, len(users))
	for _, user := range users {
		userModel, err := userEntityToModel(user)
		if err != nil {
			return nil, err
		}
		result = append(result, userModel)
	}
	return result, nil
}
