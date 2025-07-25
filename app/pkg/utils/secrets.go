package utils

import (
	"encoding/json"
	"fmt"
)

type Secret[T any] struct {
	Value T
}

func NewSecret[T any](value T) Secret[T] {
	return Secret[T]{Value: value}
}

// marshal and unmarshal Secret
func (s Secret[T]) MarshalJSON() ([]byte, error) {
	return []byte(`"***"`), nil
}
func (s *Secret[T]) UnmarshalJSON(b []byte) error {
	var val T
	var err error

	if err = json.Unmarshal(b, &val); err != nil {
		return fmt.Errorf("unmarshal user type: %w", err)
	}

	return nil
}
