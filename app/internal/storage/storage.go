package storage

import (
	"context"

	"github.com/IvLaptev/chartdb-back/internal/model"
)

type Storage interface {
	DoInTransaction(ctx context.Context, f func(ctx context.Context) error) error

	Diagram() DiagramRepository
	User() UserRepository
	UserConfirmation() UserConfirmationRepository
}

type DiagramRepository interface {
	// Supported options: [WithLock]
	GetDiagramByID(ctx context.Context, rowPolicy RowPolicy, id model.DiagramID, opts ...RequestOption) (*model.Diagram, error)
	GetAllDiagrams(ctx context.Context, rowPolicy RowPolicy, filter []*model.FilterTerm, page *model.CurrentPage) (*model.DiagramList, error)

	CreateDiagram(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error)
	PatchDiagram(ctx context.Context, params *PatchDiagramParams) (*model.Diagram, error)
	DeleteDiagram(ctx context.Context, id model.DiagramID) (*model.Diagram, error)
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id model.UserID) (*model.User, error)
	// Supported options: [WithLock]
	GetAllUsers(ctx context.Context, filter []*model.FilterTerm, options ...RequestOption) ([]*model.User, error)

	CreateUser(ctx context.Context, params *CreateUserParams) (*model.User, error)
	PatchUser(ctx context.Context, params *PatchUserParams) (*model.User, error)
	DeleteUser(ctx context.Context, userID model.UserID) (*model.User, error)
}

type UserConfirmationRepository interface {
	GetUserConfirmationByID(ctx context.Context, id model.UserConfirmationID) (*model.UserConfirmation, error)
	GetAllUserConfirmation(ctx context.Context, filter []*model.FilterTerm) ([]*model.UserConfirmation, error)

	CreateUserConfirmation(ctx context.Context, params *CreateUserConfirmationParams) (*model.UserConfirmation, error)
}
