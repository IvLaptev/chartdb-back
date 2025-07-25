package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const diagramTable = "diagrams"

var (
	diagramFields = []string{fieldID, fieldUserID, fieldClientDiagramID, fieldCode,
		fieldObjectStorageKey, fieldName, fieldTablesCount, fieldCreatedAt,
		fieldUpdatedAt, fieldDeletedAt}

	returningDiagram = returning + strings.Join(diagramFields, separator)
)

type diagramEntity struct {
	ID               model.DiagramID `db:"id"`
	UserID           model.UserID    `db:"user_id"`
	ClientDiagramID  string          `db:"client_diagram_id"`
	Code             string          `db:"code"`
	ObjectStorageKey string          `db:"object_storage_key"`
	Name             string          `db:"name"`
	TablesCount      int64           `db:"tables_count"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	DeletedAt        *time.Time      `db:"deleted_at"`
}

func (s *Storage) GetDiagramByID(ctx context.Context, rowPolicy storage.RowPolicy, id model.DiagramID, opts ...storage.RequestOption) (*model.Diagram, error) {
	options := storage.NewOptions(opts)

	query := sq.Select(diagramFields...).
		From(diagramTable).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: id.String()}).
		PlaceholderFormat(sq.Dollar)

	query, err := filterQuery(query, diagramTable, rowPolicy.GetFilter())
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}

	if options.UseLock {
		query = useLock(query, diagramTable)
	}

	sql, args := query.MustSql()

	var diagramEntity diagramEntity
	err = sqlx.GetContext(ctx, s.DB(ctx), &diagramEntity, sql, args...)
	if err != nil {
		return nil, formatError(err)
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
		return nil, formatError(err)
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
			params.Name,
			params.TablesCount,

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
		return nil, formatError(err)
	}

	return diagramEntityToModel(&diagramEntity), nil
}

func (s *Storage) PatchDiagram(ctx context.Context, params *storage.PatchDiagramParams) (*model.Diagram, error) {
	now := time.Now()

	query := sq.Update(diagramTable).
		SetMap(map[string]interface{}{
			fieldUpdatedAt: now,
		}).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: params.ID.String()}).
		Suffix(returningDiagram).
		PlaceholderFormat(sq.Dollar)

	query = patchQueryOptional(query, fieldName, params.Name)
	query = patchQueryOptional(query, fieldTablesCount, params.TablesCount)
	query = patchQueryOptional(query, fieldObjectStorageKey, params.ObjectStorageKey)

	sql, args := query.MustSql()

	var diagramEntity diagramEntity
	err := sqlx.GetContext(ctx, s.DB(ctx), &diagramEntity, sql, args...)
	if err != nil {
		return nil, err
	}

	return diagramEntityToModel(&diagramEntity), nil
}

func (s *Storage) DeleteDiagram(ctx context.Context, id model.DiagramID) (*model.Diagram, error) {
	sql, args := sq.Update(diagramTable).
		SetMap(map[string]interface{}{
			fieldDeletedAt: time.Now(),
		}).
		Where(sq.Eq{fieldDeletedAt: nil, fieldID: id.String()}).
		PlaceholderFormat(sq.Dollar).
		Suffix(returningDiagram).
		MustSql()

	var diagramEntity diagramEntity
	err := sqlx.GetContext(ctx, s.DB(ctx), &diagramEntity, sql, args...)
	if err != nil {
		return nil, formatError(err)
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
		Name:             entity.Name,
		TablesCount:      entity.TablesCount,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
		Content:          utils.NewSecret[*string](nil),
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
