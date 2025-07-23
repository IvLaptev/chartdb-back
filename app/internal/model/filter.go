package model

import "fmt"

type FilterTerm struct {
	Key       TermKey
	Value     any
	Operation FilterOperation
}

const (
	Unspecified = "UNSPECIFIED"

	TermID           = "id"
	TermCode         = "code"
	TermUserID       = "user_id"
	TermLogin        = "login"
	TermPasswordHash = "password_hash"
	TermType         = "type"
	TermConfirmedAt  = "confirmed_at"
)

type TermKey int64

const (
	TermKeyUnspecified TermKey = iota
	TermKeyID
	TermKeyCode
	TermKeyUserID
	TermKeyLogin
	TermKeyPasswordHash
	TermKeyType
	TermKeyConfirmedAt
)

func (k TermKey) String() string {
	switch k {
	case TermKeyID:
		return TermID
	case TermKeyCode:
		return TermCode
	case TermKeyUserID:
		return TermUserID
	case TermKeyLogin:
		return TermLogin
	case TermKeyPasswordHash:
		return TermPasswordHash
	case TermKeyType:
		return TermType
	case TermKeyConfirmedAt:
		return TermConfirmedAt
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
	case TermLogin:
		return TermKeyLogin, nil
	case TermPasswordHash:
		return TermKeyPasswordHash, nil
	case TermType:
		return TermKeyType, nil
	case TermConfirmedAt:
		return TermKeyConfirmedAt, nil
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
