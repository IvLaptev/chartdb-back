package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	chartdbapi "github.com/IvLaptev/chartdb-back/api/chartdb/v1"
	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/service/diagram"
)

var diagramAllowedTermKeys = map[model.TermKey]struct{}{}

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

func (h *DiagramHandler) List(ctx context.Context, req *chartdbapi.ListDiagramsRequest) (*chartdbapi.ListDiagramsResponse, error) {
	filter, err := MakeFilter(diagramAllowedTermKeys, req.Filter)
	if err != nil {
		return nil, fmt.Errorf("make filter: %w", err)
	}

	diagrams, err := h.DiagramService.ListDiagrams(ctx, &diagram.ListDiagramsParams{
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("list diagrams: %w", err)
	}

	return &chartdbapi.ListDiagramsResponse{
		Diagrams: makeDiagramMetadataList(diagrams),
	}, nil
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
		Name:            req.Name,
		TablesCount:     req.TablesCount,
	})
	if err != nil {
		return nil, fmt.Errorf("create diagram: %w", err)
	}

	return diagramMetadataToPB(diagramModel), nil
}

func (h *DiagramHandler) Update(ctx context.Context, req *chartdbapi.UpdateDiagramRequest) (*chartdbapi.DiagramMetadata, error) {
	paths, err := ExtractPaths(req.UpdateMask, req)
	if err != nil {
		return nil, fmt.Errorf("extract paths: %w", err)
	}

	patchDiagramParams := &diagram.PatchDiagramParams{
		ID:          model.DiagramID(req.Id),
		Content:     ApplyFieldOptional(req.Fields.Content, "content", paths),
		Name:        ApplyFieldOptional(req.Fields.Name, "name", paths),
		TablesCount: ApplyFieldOptional(req.Fields.TablesCount, "tables_count", paths),
	}

	diagramModel, err := h.DiagramService.PatchDiagram(ctx, patchDiagramParams)
	if err != nil {
		return nil, fmt.Errorf("patch diagram: %w", err)
	}

	return diagramMetadataToPB(diagramModel), nil
}

func (h *DiagramHandler) Delete(ctx context.Context, req *chartdbapi.DeleteDiagramRequest) (*emptypb.Empty, error) {
	_, err := h.DiagramService.DeleteDiagram(ctx, &diagram.DeleteDiagramParams{
		ID: model.DiagramID(req.Id),
	})
	if err != nil {
		return nil, fmt.Errorf("delete diagram: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func diagramMetadataToPB(diagramModel *model.Diagram) *chartdbapi.DiagramMetadata {
	return &chartdbapi.DiagramMetadata{
		Id:              diagramModel.ID.String(),
		UserId:          diagramModel.UserID.String(),
		ClientDiagramId: diagramModel.ClientDiagramID,
		Code:            diagramModel.Code,
		Name:            diagramModel.Name,
		TablesCount:     diagramModel.TablesCount,
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

func makeDiagramMetadataList(diagrams []*model.Diagram) []*chartdbapi.DiagramMetadata {
	diagramMetadatas := make([]*chartdbapi.DiagramMetadata, 0, len(diagrams))
	for _, diagram := range diagrams {
		diagramMetadata := diagramMetadataToPB(diagram)
		diagramMetadatas = append(diagramMetadatas, diagramMetadata)
	}
	return diagramMetadatas
}
