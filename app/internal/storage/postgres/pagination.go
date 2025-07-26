package postgres

import (
	"fmt"

	"github.com/IvLaptev/chartdb-back/internal/model"
	sq "github.com/Masterminds/squirrel"
)

func pageQuery(query sq.SelectBuilder, table string, page *model.CurrentPage) (sq.SelectBuilder, error) {
	if page == nil {
		return query, nil
	}

	field, err := resolveOrderByField(page.OrderBy)
	if err != nil {
		return query, fmt.Errorf("resolve order by field: %w", err)
	}
	field = tableField(table, field)

	if page.OrderBy.LastValue() != nil {
		query = query.Where(sq.Gt{field: page.OrderBy.LastValue()})
	}
	direction, err := resolveOrderByDirection(page.OrderBy.Direction())
	if err != nil {
		return query, fmt.Errorf("resolve order by direction: %w", err)
	}

	return query.Limit(page.PageSize).OrderBy(fmt.Sprintf("%s %s", field, direction)), nil
}

func resolveOrderByField(ob model.OrderBy) (string, error) {
	switch ob.(type) {
	case model.OrderByID:
		return fieldID, nil
	case model.OrderByCreatedAt:
		return fieldCreatedAt, nil
	case model.OrderByUpdatedAt:
		return fieldUpdatedAt, nil
	default:
		return "", fmt.Errorf("unsupported orderBy type: %T", ob)
	}
}

func resolveOrderByDirection(direction model.OrderByDirection) (string, error) {
	switch direction {
	case model.OrderByAsc:
		return asc, nil
	case model.OrderByDesc:
		return desc, nil
	default:
		return "", fmt.Errorf("unsupported orderByDirection type: %d", direction)
	}
}
