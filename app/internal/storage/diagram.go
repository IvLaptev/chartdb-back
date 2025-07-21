package storage

import "github.com/IvLaptev/chartdb-back/internal/model"

type CreateDiagramParams struct {
	ID               model.DiagramID
	ClientDiagramID  string
	Code             string
	UserID           model.UserID
	ObjectStorageKey string
}
