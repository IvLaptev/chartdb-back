package utils

import "errors"

// ErrMissing happens when optional has no value set
var ErrMissing = errors.New("optional has no value")

// Optional holds T value
type Optional[T any] struct {
	Value T
	Valid bool
}

// NewOptional creates new Optional[T] instance
func NewOptional[T any](v T) Optional[T] {
	return Optional[T]{Value: v, Valid: true}
}

// NewEmptyOptional creates new empty Optional[T] instance
func NewEmptyOptional[T any]() Optional[T] {
	var empty T

	return Optional[T]{Value: empty, Valid: false}
}

// From creates new Optional[T] instance
func (o Optional[T]) From(v T) Optional[T] {
	var res Optional[T]
	res.Set(v)

	return res
}

// Set value
func (o *Optional[T]) Set(v T) {
	o.Value = v
	o.Valid = true
}

// Get value if its set or return an error
func (o Optional[T]) Get() (T, error) {
	if !o.Valid {
		return *new(T), ErrMissing
	}

	return o.Value, nil
}

func OptionalFromPointer[T any](v *T) Optional[T] {
	if v == nil {
		return NewEmptyOptional[T]()
	}
	return NewOptional(*v)
}
