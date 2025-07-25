package model

import (
	"time"

	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

type DiagramID string

func (i DiagramID) String() string {
	return string(i)
}

type Diagram struct {
	ID               DiagramID
	ClientDiagramID  string
	Code             string
	UserID           UserID
	ObjectStorageKey string
	Name             string
	TablesCount      int64
	Content          utils.Secret[*string]
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
