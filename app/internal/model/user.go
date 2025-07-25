package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

const (
	Guest   = "GUEST"
	Student = "STUDENT"
	Teacher = "TEACHER"
	Admin   = "ADMIN"
)

type UserID string

func (i UserID) String() string {
	return string(i)
}

type UserType int64

const (
	UserTypeUnspecified UserType = iota
	UserTypeGuest
	UserTypeStudent
	UserTypeTeacher
	UserTypeAdmin
)

func (t UserType) String() string {
	switch t {
	case UserTypeGuest:
		return Guest
	case UserTypeStudent:
		return Student
	case UserTypeTeacher:
		return Teacher
	case UserTypeAdmin:
		return Admin
	default:
		return Unspecified
	}
}

func UserTypeFromString(s string) (UserType, error) {
	switch s {
	case Guest:
		return UserTypeGuest, nil
	case Student:
		return UserTypeStudent, nil
	case Teacher:
		return UserTypeTeacher, nil
	case Admin:
		return UserTypeAdmin, nil
	default:
		return UserTypeUnspecified, fmt.Errorf("invalid user type: %s", s)
	}
}

func (t UserType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *UserType) UnmarshalJSON(b []byte) error {
	var s string
	var err error

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unmarshal user type: %w", err)
	}

	*t, err = UserTypeFromString(s)
	if err != nil {
		return fmt.Errorf("user type from string: %w", err)
	}

	return nil
}

type User struct {
	ID           UserID                `json:"id"`
	Login        string                `json:"login"`
	PasswordHash utils.Secret[*string] `json:"password_hash"`
	Type         UserType              `json:"type"`
	ConfirmedAt  *time.Time            `json:"confirmed_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

type UserConfirmationID string

func (i UserConfirmationID) String() string {
	return string(i)
}

type UserConfirmation struct {
	ID        UserConfirmationID `json:"id"`
	UserID    UserID             `json:"user_id"`
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
}

type UserToken struct {
	Value  string `json:"value"`
	UserID UserID `json:"user_id"`
}
