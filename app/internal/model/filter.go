package model

import "fmt"

type FilterTerm struct {
	Key       TermKey
	Value     any
	Operation FilterOperation
}

const (
	Unspecified = "UNSPECIFIED"

	TermID     = "id"
	TermCode   = "code"
	TermUserID = "user_id"
)

type TermKey int64

const (
	TermKeyUnspecified TermKey = iota
	TermKeyID
	TermKeyCode
	TermKeyUserID
)

func (k TermKey) String() string {
	switch k {
	case TermKeyID:
		return TermID
	case TermKeyCode:
		return TermCode
	case TermKeyUserID:
		return TermUserID
	default:
		return Unspecified
	}
}

func TermKeyFromString(str string) (TermKey, error) {
	switch str {
	case TermID:
		return TermKeyID, nil
	case TermCode:
		return TermKeyCode, nil
	case TermUserID:
		return TermKeyUserID, nil
	default:
		return 0, fmt.Errorf("invalid term key: %s", str)
	}
}

const (
	Exact     = "="
	Substring = ":"
	IsNull    = "nil"
	NotEqual  = "!="
	Less      = "<"
	More      = ">"
	Contains  = "@>"
)

type FilterOperation int64

const (
	FilterOperationUnspecified FilterOperation = iota
	FilterOperationExact
	FilterOperationSubstring
	FilterOperationIsNil
	FilterOperationNotEqual
	FilterOperationLess
	FilterOperationMore
	FilterContains
)

func (o FilterOperation) String() string {
	switch o {
	case FilterOperationExact:
		return Exact
	case FilterOperationSubstring:
		return Substring
	case FilterOperationIsNil:
		return IsNull
	case FilterOperationNotEqual:
		return NotEqual
	case FilterOperationLess:
		return Less
	case FilterOperationMore:
		return More
	case FilterContains:
		return Contains
	}

	return Unspecified
}
