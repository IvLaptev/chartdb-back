package storage

import (
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

type CreateUserParams struct {
	ID           model.UserID
	Login        string
	PasswordHash *string
	Type         model.UserType
	ConfirmedAt  *time.Time
}

type PatchUserParams struct {
	ID          model.UserID
	ConfirmedAt utils.Optional[*time.Time]
}
