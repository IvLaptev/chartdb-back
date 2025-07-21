package storage

import (
	"context"
	"fmt"

	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
)

type RowPolicy interface {
	GetFilter() []*model.FilterTerm
}

type RowPolicyUserID struct {
	UserID model.UserID
}

func (s *RowPolicyUserID) GetFilter() []*model.FilterTerm {
	return []*model.FilterTerm{
		{
			Key:       model.TermKeyUserID,
			Value:     s.UserID,
			Operation: model.FilterOperationExact,
		},
	}
}

type RowPolicyBackground struct{}

func (s *RowPolicyBackground) GetFilter() []*model.FilterTerm {
	return []*model.FilterTerm{}
}

func RowPolicyFromContext(ctx context.Context) (RowPolicy, error) {
	subject, err := auth.GetSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}

	return &RowPolicyUserID{
		UserID: subject.UserID,
	}, nil
}
