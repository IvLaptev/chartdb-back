package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	sq "github.com/Masterminds/squirrel"
)

const diagramTable = "diagrams"

var (
	diagramFields = []string{fieldID, fieldUserID, fieldClientDiagramID, fieldCode,
		fieldObjectStorageKey, fieldCreatedAt, fieldDeletedAt}
)

type Diagram struct {
	ID               model.DiagramID `db:"id"`
	UserID           model.UserID    `db:"user_id"`
	ClientDiagramID  string          `db:"client_diagram_id"`
	Code             string          `db:"code"`
	ObjectStorageKey string          `db:"object_storage_key"`
	CreatedAt        time.Time       `db:"created_at"`
	DeletedAt        *time.Time      `db:"deleted_at"`
}

func (s *Storage) GetAllDiagrams(ctx context.Context, filter []*model.FilterTerm) ([]*model.Diagram, error) {
	query := sq.Select(diagramFields...).
		Where(sq.Eq{fieldDeletedAt: nil}).
		From(diagramTable).
		PlaceholderFormat(sq.Dollar)

	query, err := filterQuery(query, diagramTable, filter)
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}

	sql, args := query.MustSql()

	var diagramEntities []*Diagram
	err = s.db.SelectContext(ctx, &diagramEntities, sql, args...)
	if err != nil {
		return nil, err
	}

	return makeDiagramList(diagramEntities), nil
}

func (s *Storage) CreateDiagram(ctx context.Context, diagram *model.Diagram) error {
	now := time.Now()

	sql, args := sq.
		Insert(diagramTable).
		Columns(diagramFields...).
		Values(
			diagram.ID,
			diagram.UserID,
			diagram.ClientDiagramID,
			diagram.Code,
			diagram.ObjectStorageKey,

			now,
			nil,
		).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	_, err := s.db.ExecContext(ctx, sql, args...)
	return err
}

func diagramEntityToModel(entity *Diagram) *model.Diagram {
	return &model.Diagram{
		ID:               entity.ID,
		UserID:           entity.UserID,
		ClientDiagramID:  entity.ClientDiagramID,
		Code:             entity.Code,
		ObjectStorageKey: entity.ObjectStorageKey,
		Content:          nil,
	}
}

func makeDiagramList(entities []*Diagram) []*model.Diagram {
	diagrams := make([]*model.Diagram, 0, len(entities))
	for _, entity := range entities {
		diagramModel := diagramEntityToModel(entity)
		diagrams = append(diagrams, diagramModel)
	}

	return diagrams
}
