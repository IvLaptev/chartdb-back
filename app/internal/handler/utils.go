package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/IvLaptev/chartdb-back/internal/model"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var allowedFilterOperations = []model.FilterOperation{
	model.FilterOperationExact,
}

// pathExists determines if a given path or any of its prefixes exist in the specified paths map.
func pathExists(path string, paths map[string]struct{}) bool {
	if _, exists := paths[path]; exists {
		return true
	}

	components := strings.Split(path, ".")

	for i := 1; i < len(components); i++ {
		prefix := strings.Join(components[:i], ".")
		if _, exists := paths[prefix]; exists {
			return true
		}
	}

	return false
}

func ApplyFieldOptional[T any](value T, path string, paths map[string]struct{}) utils.Optional[T] {
	if pathExists(path, paths) {
		return utils.NewOptional(value)
	}

	var zero T
	return utils.Optional[T]{Value: zero, Valid: false}
}

func ExtractPaths(fieldMask *fieldmaskpb.FieldMask, request proto.Message) (map[string]struct{}, error) {
	if fieldMask == nil {
		return nil, xerrors.WrapInvalidArgument(errors.New("empty field mask"))
	}

	fieldMask.Normalize()

	return ConvertToSet(fieldMask.GetPaths()), nil
}

func ConvertToSet(paths []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, path := range paths {
		result[path] = struct{}{}
	}

	return result
}

func MakeFilter(allowedKeys map[model.TermKey]struct{}, filter string) ([]*model.FilterTerm, error) {
	if filter == "" {
		return nil, nil
	}

	terms, err := parseFilter(allowedFilterOperations, allowedKeys, filter)
	if err != nil {
		return nil, fmt.Errorf("parse filter: %w", err)
	}

	return terms, nil
}

func parseFilter(allowedOperations []model.FilterOperation, allowedKeys map[model.TermKey]struct{}, filter string) ([]*model.FilterTerm, error) {
	terms := splitTerms(filter)

	result := make([]*model.FilterTerm, 0, len(terms))
	for _, termString := range terms {
		term, err := parseTerm(termString, allowedOperations)
		if err != nil {
			return nil, fmt.Errorf("parse term: %w", err)
		}
		if _, ok := allowedKeys[term.Key]; !ok {
			return nil, fmt.Errorf("unsupported term key: %s", term.Key)
		}

		result = append(result, term)
	}

	return result, nil
}

func splitTerms(filter string) []string {
	filter = strings.TrimSpace(filter)
	filter = strings.ReplaceAll(filter, " and", " AND")

	split := strings.Split(filter, " AND")
	result := make([]string, 0, len(split))
	for _, term := range split {
		result = append(result, strings.TrimSpace(term))
	}

	return result
}

func parseTerm(term string, allowedOperations []model.FilterOperation) (*model.FilterTerm, error) {
	for _, op := range allowedOperations {
		if strings.Contains(term, op.String()) {
			split := strings.Split(term, op.String())
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid term format, op: %s, found parts: %d", op, len(split))
			}
			key := strings.TrimSpace(split[0])
			camelCaseKey := snakeCaseToCamelCase(key)
			termKey, err := model.TermKeyFromString(camelCaseKey)
			if err != nil {
				return nil, fmt.Errorf("parse term key: %w", err)
			}

			value := strings.TrimSpace(split[1])
			termValue := strings.TrimFunc(value, func(r rune) bool {
				switch r {
				case '"':
					return true
				case '\'':
					return true
				}

				return false
			})

			return &model.FilterTerm{
				Key:       termKey,
				Value:     termValue,
				Operation: op,
			}, nil
		}
	}

	return nil, fmt.Errorf("can't find any of allowed operators: %v", allowedOperations)
}

func snakeCaseToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = capitalize(parts[i])
	}
	return strings.Join(parts, "")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToTitle(s[:1]) + s[1:]
}
