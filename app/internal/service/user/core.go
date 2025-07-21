package user

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"

	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
)

var (
	ErrInvalidUserID = errors.New("invalid user id")
)

type Service interface {
	Authenticate(ctx context.Context, token string) (context.Context, error)
}

type ServiceImpl struct {
	Logger *slog.Logger
}

func (s *ServiceImpl) Authenticate(ctx context.Context, token string) (context.Context, error) {
	userID, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, xerrors.WrapUnauthenticated(ErrInvalidUserID)
	}

	ctx = auth.SetSubject(ctx, &auth.Subject{
		UserID: model.UserID(string(userID)),
	})

	return ctx, nil
}

func NewService(logger *slog.Logger) Service {
	return &ServiceImpl{
		Logger: logger,
	}
}
