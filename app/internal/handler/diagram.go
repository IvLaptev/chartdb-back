package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	chartdbapi "github.com/IvLaptev/chartdb-back/api/chartdb/v1"
	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/service/diagram"
)

type DiagramHandler struct {
	chartdbapi.UnimplementedDiagramServiceServer

	Logger         *slog.Logger
	DiagramService diagram.Service
}

func (h *DiagramHandler) Get(ctx context.Context, req *chartdbapi.GetDiagramRequest) (*chartdbapi.Diagram, error) {
	diagramModel, err := h.DiagramService.GetDiagram(ctx, &diagram.GetDiagramParams{
		Identifier: strings.ToLower(req.Identifier),
	})
	if err != nil {
		return nil, fmt.Errorf("get diagram: %w", err)
	}

	return diagramToPB(diagramModel)
}

func (h *DiagramHandler) Create(ctx context.Context, req *chartdbapi.CreateDiagramRequest) (*chartdbapi.DiagramMetadata, error) {
	subject, err := auth.GetSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}

	diagramModel, err := h.DiagramService.CreateDiagram(ctx, &diagram.CreateDiagramParams{
		ClientDiagramID: req.ClientDiagramId,
		UserID:          subject.UserID,
		Content:         req.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("create diagram: %w", err)
	}

	return diagramMetadataToPB(diagramModel), nil
}

func diagramMetadataToPB(diagramModel *model.Diagram) *chartdbapi.DiagramMetadata {
	return &chartdbapi.DiagramMetadata{
		Id:              diagramModel.ID.String(),
		UserId:          diagramModel.UserID.String(),
		ClientDiagramId: diagramModel.ClientDiagramID,
		Code:            diagramModel.Code,
		CreatedAt:       timestamppb.New(diagramModel.CreatedAt),
		UpdatedAt:       timestamppb.New(diagramModel.UpdatedAt),
	}
}

func diagramToPB(diagramModel *model.Diagram) (*chartdbapi.Diagram, error) {
	content := ""
	if diagramModel.Content != nil {
		content = *diagramModel.Content
	}

	return &chartdbapi.Diagram{
		Metadata: diagramMetadataToPB(diagramModel),
		Content:  content,
	}, nil
}
