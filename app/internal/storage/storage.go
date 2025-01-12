package storage

import (
	"context"

	"github.com/IvLaptev/chartdb-back/internal/model"
)

type Storage interface {
	DoInTransaction(ctx context.Context, f func(ctx context.Context) error) error

	Diagram() DiagramRepository
}

type DiagramRepository interface {
	GetAllDiagrams(ctx context.Context, filter []*model.FilterTerm) ([]*model.Diagram, error)

	CreateDiagram(ctx context.Context, diagram *model.Diagram) error
}
