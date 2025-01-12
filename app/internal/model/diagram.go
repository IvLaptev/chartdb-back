package model

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
}
