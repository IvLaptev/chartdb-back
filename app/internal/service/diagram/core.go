package diagram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/internal/utils"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
)

const (
	diagramIDLength        int64 = 10
	codeLength             int64 = 4
	objectStorageKeyLength int64 = 20
)

var (
	ErrDiagramNotFound        = errors.New("diagram not found")
	ErrDiagramContentNotFound = errors.New("diagram content not found")
)

type Service interface {
	GetDiagram(ctx context.Context, params *GetDiagramParams) (*model.Diagram, error)
	CreateDiagram(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error)
}

type ServiceImpl struct {
	Storage  storage.Storage
	S3Client s3client.Client
	Logger   *slog.Logger
}

type GetDiagramParams struct {
	Identifier string
}

func (s *ServiceImpl) GetDiagram(ctx context.Context, params *GetDiagramParams) (*model.Diagram, error) {
	ctxlog.Info(ctx, s.Logger, "get diagram", slog.Any("params", params))

	rowPolicy, err := storage.RowPolicyFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("row polycy from context: %w", err)
	}

	var diagramModel *model.Diagram
	diagramList, err := s.Storage.Diagram().GetAllDiagrams(ctx, rowPolicy, []*model.FilterTerm{
		{
			Key:       model.TermKeyID,
			Value:     params.Identifier,
			Operation: model.FilterOperationExact,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get all diagrams by id: %w", err)
	}

	if len(diagramList) == 0 {
		diagramList, err = s.Storage.Diagram().GetAllDiagrams(ctx, rowPolicy, []*model.FilterTerm{
			{
				Key:       model.TermKeyCode,
				Value:     params.Identifier,
				Operation: model.FilterOperationExact,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("get all diagrams by code: %w", err)
		}
		if len(diagramList) == 0 {
			return nil, xerrors.WrapNotFound(ErrDiagramNotFound)
		}
	}

	diagramModel = diagramList[0]

	content, err := s.S3Client.GetContent(ctx, diagramModel.ObjectStorageKey)
	if err != nil {
		if errors.Is(err, s3client.ErrContentNotFound) {
			return nil, xerrors.WrapNotFound(ErrDiagramContentNotFound)
		}
		return nil, fmt.Errorf("get content: %w", err)
	}

	diagramModel.Content = &content

	return diagramModel, nil
}

type CreateDiagramParams struct {
	ClientDiagramID string
	UserID          model.UserID
	Content         string
}

func (s *ServiceImpl) CreateDiagram(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error) {
	ctxlog.Info(ctx, s.Logger, "create diagram", slog.Any("params", params))

	diagramID, err := utils.GenerateID(diagramIDLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	code, err := utils.GenerateID(codeLength)
	if err != nil {
		return nil, fmt.Errorf("generate id (code): %w", err)
	}

	objStorageKey, err := utils.GenerateID(objectStorageKeyLength)
	if err != nil {
		return nil, fmt.Errorf("generate id (storage key): %w", err)
	}

	var diagramModel *model.Diagram
	err = s.Storage.DoInTransaction(ctx, func(ctx context.Context) error {
		diagramModel, err = s.Storage.Diagram().CreateDiagram(ctx, &storage.CreateDiagramParams{
			ID:               model.DiagramID(diagramID),
			ClientDiagramID:  params.ClientDiagramID,
			Code:             code,
			UserID:           params.UserID,
			ObjectStorageKey: objStorageKey,
		})
		if err != nil {
			return fmt.Errorf("create diagram: %w", err)
		}

		err = s.S3Client.SaveContent(ctx, diagramModel.ObjectStorageKey, params.Content)
		if err != nil {
			return fmt.Errorf("save content: %w", err)
		}

		diagramModel.Content = &params.Content

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't create diagram: %w", err)
	}

	return diagramModel, nil
}

func NewService(logger *slog.Logger, storage storage.Storage, s3Client s3client.Client) *ServiceImpl {
	return &ServiceImpl{
		Logger:   logger.With("name", "service/diagram"),
		Storage:  storage,
		S3Client: s3Client,
	}
}
