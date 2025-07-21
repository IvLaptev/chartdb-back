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
	GetDiagramByID(ctx context.Context, rowPolicy RowPolicy, id model.DiagramID) (*model.Diagram, error)
	GetAllDiagrams(ctx context.Context, rowPolicy RowPolicy, filter []*model.FilterTerm) ([]*model.Diagram, error)

	CreateDiagram(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error)
}
