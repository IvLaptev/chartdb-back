package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	defaultPageSize = 100

	OrderByIDName        = "id"
	OrderByCreatedAtName = "createdAt"
	OrderByUpdateAtName  = "updatedAt"

	asc  = "asc"
	desc = "desc"
)

type OrderByDirection uint8

const (
	OrderByAsc OrderByDirection = iota
	OrderByDesc
)

type OrderBy interface {
	FieldName() string
	LastValue() *string
	Direction() OrderByDirection
	withDirection(OrderByDirection) OrderBy
}

type OrderByID struct {
	LastID           *string
	OrderByDirection OrderByDirection
}

func (o OrderByID) FieldName() string { return OrderByIDName }

func (o OrderByID) LastValue() *string {
	return o.LastID
}

func (o OrderByID) Direction() OrderByDirection {
	return o.OrderByDirection
}

func (o OrderByID) withDirection(direction OrderByDirection) OrderBy {
	o.OrderByDirection = direction
	return o
}

func ResolveOrderByDirection(direction string) (OrderByDirection, error) {
	switch direction {
	case asc:
		return OrderByAsc, nil
	case desc:
		return OrderByDesc, nil
	default:
		return 0, fmt.Errorf("invalid orderBy direction: %s", direction)
	}
}

type OrderByCreatedAt struct {
	LastTime         *string
	OrderByDirection OrderByDirection
}

func (o OrderByCreatedAt) FieldName() string { return OrderByCreatedAtName }

func (o OrderByCreatedAt) LastValue() *string {
	return o.LastTime
}

func (o OrderByCreatedAt) Direction() OrderByDirection {
	return o.OrderByDirection
}

func (o OrderByCreatedAt) withDirection(direction OrderByDirection) OrderBy {
	o.OrderByDirection = direction
	return o
}

type OrderByUpdatedAt struct {
	LastTime         *string
	OrderByDirection OrderByDirection
}

func (o OrderByUpdatedAt) FieldName() string { return OrderByUpdateAtName }

func (o OrderByUpdatedAt) LastValue() *string {
	return o.LastTime
}

func (o OrderByUpdatedAt) Direction() OrderByDirection {
	return o.OrderByDirection
}

func (o OrderByUpdatedAt) withDirection(direction OrderByDirection) OrderBy {
	o.OrderByDirection = direction
	return o
}

type CurrentPage struct {
	PageSize uint64
	OrderBy  OrderBy
}

func (p *CurrentPage) NextPage(orderBy OrderBy, resultSize int) (*NextPage, error) {
	if reflect.TypeOf(p.OrderBy) != reflect.TypeOf(orderBy) {
		return nil, fmt.Errorf("invalid orderBy type, expected: %T, actual: %T", p.OrderBy, orderBy)
	}
	if uint64(resultSize) < p.PageSize {
		return nil, nil
	}

	return &NextPage{
		pageSize: p.PageSize,
		orderBy:  orderBy,
	}, nil
}

type pageOptions struct {
	direction OrderByDirection
}

type PageOption func(*pageOptions)

func WithDirection(direction OrderByDirection) PageOption {
	return func(options *pageOptions) {
		options.direction = direction
	}
}

func NewPage[T OrderBy](pageSize int64, pageTokenString string, opts ...PageOption) (*CurrentPage, error) {
	if pageSize < 0 {
		return nil, fmt.Errorf("invalid page size: %d", pageSize)
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	options := &pageOptions{
		direction: OrderByAsc,
	}
	for _, o := range opts {
		o(options)
	}

	page := &CurrentPage{
		PageSize: uint64(pageSize),
		OrderBy:  *new(T),
	}
	page.OrderBy = page.OrderBy.withDirection(options.direction)
	if reflect.ValueOf(page.OrderBy).Kind() == reflect.Ptr {
		return nil, fmt.Errorf("invalid page type, found pointer")
	}

	if pageTokenString != "" {
		pt, err := decodePageToken(pageTokenString)
		if err != nil {
			return nil, fmt.Errorf("decode page token: %w", err)
		}
		orderBy, err := makeOrderBy(pt)
		if err != nil {
			return nil, fmt.Errorf("make orderBy: %w", err)
		}
		if reflect.TypeOf(page.OrderBy) != reflect.TypeOf(orderBy) {
			return nil, fmt.Errorf("invalid orderBy type, requested: %T, but decoded: %T", page.OrderBy, orderBy)
		}
		page.OrderBy = orderBy
	}

	return page, nil
}

type NextPage struct {
	pageSize uint64
	orderBy  OrderBy
}

func (p *NextPage) Token() (string, error) {
	if p == nil {
		return "", nil
	}

	return encodePageToken(&pageToken{
		Field:     p.orderBy.FieldName(),
		LastValue: p.orderBy.LastValue(),
		Direction: p.orderBy.Direction(),
	})
}

type pageToken struct {
	Field     string           `json:"field"`
	LastValue *string          `json:"last_value"`
	Direction OrderByDirection `json:"direction"`
}

func encodePageToken(pt *pageToken) (string, error) {
	marshal, err := json.Marshal(pt)
	if err != nil {
		return "", fmt.Errorf("json marshal page token: %w", err)
	}

	return base64.RawStdEncoding.EncodeToString(marshal), nil
}

func decodePageToken(pageTokenString string) (*pageToken, error) {
	decodedString, err := base64.RawStdEncoding.DecodeString(pageTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid page token format: %w", err)
	}

	var pt pageToken
	err = json.Unmarshal(decodedString, &pt)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling page token: %w", err)
	}

	return &pt, nil
}

func makeOrderBy(pt *pageToken) (OrderBy, error) {
	if pt == nil {
		return nil, fmt.Errorf("pageToken is nil")
	}

	switch pt.Field {
	case OrderByIDName:
		return OrderByID{
			LastID:           pt.LastValue,
			OrderByDirection: pt.Direction,
		}, nil
	case OrderByCreatedAtName:
		return OrderByCreatedAt{
			LastTime:         pt.LastValue,
			OrderByDirection: pt.Direction,
		}, nil
	case OrderByUpdateAtName:
		return OrderByUpdatedAt{
			LastTime:         pt.LastValue,
			OrderByDirection: pt.Direction,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported orderBy type: %s", pt.Field)
	}
}
