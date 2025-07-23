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

const diagramTable = "diagrams"

var (
	diagramFields = []string{fieldID, fieldUserID, fieldClientDiagramID, fieldCode,
		fieldObjectStorageKey, fieldCreatedAt, fieldUpdatedAt, fieldDeletedAt}

	returningDiagram = returning + strings.Join(diagramFields, separator)
)

type diagramEntity struct {
	ID               model.DiagramID `db:"id"`
	UserID           model.UserID    `db:"user_id"`
	ClientDiagramID  string          `db:"client_diagram_id"`
	Code             string          `db:"code"`
	ObjectStorageKey string          `db:"object_storage_key"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	DeletedAt        *time.Time      `db:"deleted_at"`
}

func (s *Storage) GetDiagramByID(ctx context.Context, rowPolicy storage.RowPolicy, id model.DiagramID) (*model.Diagram, error) {
	query := sq.Select(diagramFields...).
		From(diagramTable).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: id.String()}).
		PlaceholderFormat(sq.Dollar)

	query, err := filterQuery(query, diagramTable, rowPolicy.GetFilter())
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}

	sql, args := query.MustSql()

	var diagramEntity diagramEntity
	err = sqlx.GetContext(ctx, s.DB(ctx), &diagramEntity, sql, args...)
	if err != nil {
		return nil, err
	}

	return diagramEntityToModel(&diagramEntity), nil
}

func (s *Storage) GetAllDiagrams(ctx context.Context, rowPolicy storage.RowPolicy, filter []*model.FilterTerm) ([]*model.Diagram, error) {
	query := sq.Select(diagramFields...).
		Where(sq.Eq{fieldDeletedAt: nil}).
		From(diagramTable).
		PlaceholderFormat(sq.Dollar)

	query, err := filterQuery(query, diagramTable, append(filter, rowPolicy.GetFilter()...))
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}

	sql, args := query.MustSql()

	var diagramEntities []*diagramEntity
	err = sqlx.SelectContext(ctx, s.DB(ctx), &diagramEntities, sql, args...)
	if err != nil {
		return nil, err
	}

	return makeDiagramList(diagramEntities), nil
}

func (s *Storage) CreateDiagram(ctx context.Context, params *storage.CreateDiagramParams) (*model.Diagram, error) {
	now := time.Now()

	sql, args := sq.
		Insert(diagramTable).
		Columns(diagramFields...).
		Values(
			params.ID.String(),
			params.UserID.String(),
			params.ClientDiagramID,
			params.Code,
			params.ObjectStorageKey,

			now,
			now,
			nil,
		).
		Suffix(returningDiagram).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	var diagramEntity diagramEntity
	err := sqlx.GetContext(ctx, s.DB(ctx), &diagramEntity, sql, args...)
	if err != nil {
		return nil, err
	}

	return diagramEntityToModel(&diagramEntity), nil
}

func diagramEntityToModel(entity *diagramEntity) *model.Diagram {
	return &model.Diagram{
		ID:               entity.ID,
		UserID:           entity.UserID,
		ClientDiagramID:  entity.ClientDiagramID,
		Code:             entity.Code,
		ObjectStorageKey: entity.ObjectStorageKey,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
		Content:          nil,
	}
}

func makeDiagramList(entities []*diagramEntity) []*model.Diagram {
	diagrams := make([]*model.Diagram, 0, len(entities))
	for _, entity := range entities {
		diagramModel := diagramEntityToModel(entity)
		diagrams = append(diagrams, diagramModel)
	}

	return diagrams
}
