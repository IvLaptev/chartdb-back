package auth

import (
	"context"
	"errors"

	"github.com/IvLaptev/chartdb-back/internal/model"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
)

var (
	ErrSubjectNotFound = errors.New("subject not found")
)

type subjectKey struct{}

type Subject struct {
	UserID model.UserID
}

func SetSubject(ctx context.Context, subject *Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, subject)
}

func GetSubject(ctx context.Context) (*Subject, error) {
	if subject, ok := ctx.Value(subjectKey{}).(*Subject); ok {
		return subject, nil
	}
	return nil, xerrors.WrapUnauthenticated(ErrSubjectNotFound)
}
