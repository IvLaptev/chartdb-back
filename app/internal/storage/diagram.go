package storage

import (
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

type CreateDiagramParams struct {
	ID               model.DiagramID
	ClientDiagramID  string
	Code             string
	UserID           model.UserID
	ObjectStorageKey string
	Name             string
	TablesCount      int64
}

type PatchDiagramParams struct {
	ID model.DiagramID

	Name             utils.Optional[string]
	TablesCount      utils.Optional[int64]
	ObjectStorageKey utils.Optional[string]
}
