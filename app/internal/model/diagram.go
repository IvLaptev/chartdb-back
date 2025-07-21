package model

import "time"

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
	Content          *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
