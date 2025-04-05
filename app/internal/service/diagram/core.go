package diagram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/internal/utils"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
)

const (
	diagramIDLength        int64 = 10
	codeLength             int64 = 4
	objectStorageKeyLength int64 = 20
)

var (
	ErrDiagramNotFound = errors.New("diagram not found")
)

type Service interface {
	Load(ctx context.Context, params *LoadDiagramParams) (*model.Diagram, error)
	Create(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error)
}

type ServiceImpl struct {
	Storage  storage.Storage
	S3Client s3client.Client
	Logger   *slog.Logger
}

type LoadDiagramParams struct {
	UserID model.UserID
	Code   string
}

func (s *ServiceImpl) Load(ctx context.Context, params *LoadDiagramParams) (*model.Diagram, error) {
	diagramFilters := []*model.FilterTerm{
		{
			Key:       model.TermKeyUserID,
			Value:     params.UserID.String(),
			Operation: model.FilterOperationExact,
		},
		{
			Key:       model.TermKeyCode,
			Value:     strings.ToLower(params.Code),
			Operation: model.FilterOperationExact,
		},
	}

	diagramsList, err := s.Storage.Diagram().GetAllDiagrams(ctx, diagramFilters)
	if err != nil {
		return nil, fmt.Errorf("get all diagrams: %w", err)
	}

	if len(diagramsList) != 1 {
		return nil, ErrDiagramNotFound
	}

	diagramModel := diagramsList[0]

	content, err := s.S3Client.GetContent(ctx, diagramModel.ObjectStorageKey)
	if err != nil {
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

func (s *ServiceImpl) Create(ctx context.Context, params *CreateDiagramParams) (*model.Diagram, error) {
	diagramID, err := utils.GenerateID(diagramIDLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	code, err := utils.GenerateID(codeLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	objStorageKey, err := utils.GenerateID(objectStorageKeyLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	diagramModel := model.Diagram{
		ID:               model.DiagramID(diagramID),
		ClientDiagramID:  params.ClientDiagramID,
		Code:             code,
		UserID:           params.UserID,
		ObjectStorageKey: objStorageKey,
		Content:          &params.Content,
	}

	err = s.Storage.DoInTransaction(ctx, func(ctx context.Context) error {
		err := s.Storage.Diagram().CreateDiagram(ctx, &diagramModel)
		if err != nil {
			return fmt.Errorf("create diagram: %w", err)
		}

		err = s.S3Client.SaveContent(ctx, diagramModel.ObjectStorageKey, params.Content)
		if err != nil {
			return fmt.Errorf("save content: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create diagram: %w", err)
	}

	return &diagramModel, nil
}

func NewService(logger *slog.Logger, storage storage.Storage, s3Client s3client.Client) *ServiceImpl {
	return &ServiceImpl{
		Logger:   logger.With("name", "service/diagram"),
		Storage:  storage,
		S3Client: s3Client,
	}
}
