package postgres

import (
	"fmt"
	"strings"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	sq "github.com/Masterminds/squirrel"
)

func tableField(table, field string) string {
	return table + "." + field
}

func tableFieldIfNoTable(table, field string) string {
	if strings.Contains(field, ".") {
		return field
	}

	return tableField(table, field)
}

type filterableQueryBuilder[QT any] interface {
	sq.SelectBuilder | sq.UpdateBuilder | sq.DeleteBuilder
	Where(pred interface{}, args ...interface{}) QT
}

func filterQuery[T filterableQueryBuilder[T]](query T, table string, filter []*model.FilterTerm) (T, error) {
	for _, term := range filter {
		switch term.Operation {
		case model.FilterOperationExact:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.Eq{
				tableFieldIfNoTable(table, termField): term.Value,
			})
		case model.FilterOperationSubstring:
			stringTermValue, ok := term.Value.(string)
			if !ok {
				return query, fmt.Errorf("expected value of string type for filter key %s, but %T type found", term.Key, term.Value)
			}
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.Like{
				tableFieldIfNoTable(table, termField): "%" + stringTermValue + "%",
			})
		case model.FilterOperationIsNil:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.Eq{
				tableFieldIfNoTable(table, termField): nil,
			})
		case model.FilterOperationNotEqual:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.NotEq{
				tableFieldIfNoTable(table, termField): term.Value,
			})
		case model.FilterOperationLess:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.Lt{
				tableFieldIfNoTable(table, termField): term.Value,
			})
		case model.FilterOperationMore:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			query = query.Where(sq.Gt{
				tableFieldIfNoTable(table, termField): term.Value,
			})
		case model.FilterContains:
			termField, err := resolveTermField(term.Key)
			if err != nil {
				return query, err
			}
			switch term.Value.(type) {
			case int64, uint64, int32, uint32, int16, uint16, int8, uint8, int:
				query = query.Where(
					fmt.Sprintf("%s @> ?::bigint", tableFieldIfNoTable(table, termField)),
					term.Value,
				)
			default:
				query = query.Where(
					fmt.Sprintf("%s @> ?", tableFieldIfNoTable(table, termField)),
					term.Value,
				)
			}
		}
	}

	return query, nil
}

func resolveTermField(key model.TermKey) (string, error) {
	switch key {
	case model.TermKeyID:
		return fieldID, nil
	case model.TermKeyCode:
		return fieldCode, nil
	case model.TermKeyUserID:
		return fieldUserID, nil
	case model.TermKeyLogin:
		return fieldLogin, nil
	case model.TermKeyPasswordHash:
		return fieldPasswordHash, nil
	case model.TermKeyType:
		return fieldType, nil
	case model.TermKeyConfirmedAt:
		return fieldConfirmedAt, nil
	case model.TermKeyObjectStorageKey:
		return fieldObjectStorageKey, nil
	default:
		return "", fmt.Errorf("unsupported termKey type: %d", key)
	}
}

func patchQueryOptional[T any](
	query sq.UpdateBuilder,
	fieldName string,
	optValue utils.Optional[T],
) sq.UpdateBuilder {
	if optValue.Valid {
		query = query.Set(fieldName, optValue.Value)
	}
	return query
}
