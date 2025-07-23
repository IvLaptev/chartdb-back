package storage

import (
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
)

type CreateUserConfirmationParams struct {
	ID       model.UserConfirmationID
	UserID   model.UserID
	Duration time.Duration
}
