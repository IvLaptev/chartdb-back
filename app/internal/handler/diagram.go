package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/service/diagram"
	xhttp "github.com/IvLaptev/chartdb-back/pkg/http"
)

const (
	XUserIDHeader = "x-user-id"
)

type Diagram struct {
	DiagramService diagram.Service
}

func (h *Diagram) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/{code}", h.Load)

	return r
}

type LoadDiagramRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Code   string `json:"code" validate:"len=4,required,alphanum"`
}

type LoadDiagramResponse struct {
	Content string `json:"content"`
}

func (h *Diagram) Load(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := LoadDiagramRequest{
		UserID: r.Header.Get(XUserIDHeader),
		Code:   strings.ToLower(chi.URLParam(r, "code")),
	}

	if err := validator.New().Struct(req); err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(err))
		return
	}

	userID, err := base64.StdEncoding.DecodeString(req.UserID)
	if err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(errors.New("invalid user id")))
		return
	}

	diagramModel, err := h.DiagramService.Load(ctx, &diagram.LoadDiagramParams{
		UserID: model.UserID(userID),
		Code:   req.Code,
	})
	if err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(err))
		return
	}

	render.JSON(w, r, LoadDiagramResponse{
		Content: *diagramModel.Content,
	})
}

type CreateDiagramRequest struct {
	UserID          string `json:"user_id" validate:"required"`
	ClientDiagramID string `json:"client_diagram_id" validate:"len=4,required,alphanum"`
	Content         string `json:"content" validate:"required"`
}

type CreateDiagramResponse struct {
	Code            string `json:"code"`
	ClientDiagramID string `json:"client_diagram_id"`
}

func (h *Diagram) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateDiagramRequest
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(err))
		return
	}

	req.UserID = r.Header.Get(XUserIDHeader)

	if err = validator.New().Struct(req); err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(err))
		return
	}

	userID, err := base64.StdEncoding.DecodeString(req.UserID)
	if err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(errors.New("invalid user id")))
		return
	}

	diagramModel, err := h.DiagramService.Create(ctx, &diagram.CreateDiagramParams{
		ClientDiagramID: req.ClientDiagramID,
		UserID:          model.UserID(userID),
		Content:         req.Content,
	})
	if err != nil {
		render.Render(w, r, xhttp.ErrInvalidRequest(err))
		return
	}

	render.JSON(w, r, CreateDiagramResponse{
		ClientDiagramID: diagramModel.ClientDiagramID,
		Code:            diagramModel.Code,
	})
}
