package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	chartdbapi "github.com/IvLaptev/chartdb-back/api/chartdb/v1"
	"github.com/IvLaptev/chartdb-back/internal/handler"
	"github.com/IvLaptev/chartdb-back/internal/service/diagram"
	"github.com/IvLaptev/chartdb-back/internal/service/user"
	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
	"github.com/IvLaptev/chartdb-back/internal/utils"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
	xhttp "github.com/IvLaptev/chartdb-back/pkg/http"
	"github.com/IvLaptev/chartdb-back/pkg/middleware"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type application struct {
	config *Config
	logger *slog.Logger
}

func newApplication(cfg *Config, logger *slog.Logger) application {
	return application{
		config: cfg,
		logger: logger,
	}
}

func (a *application) Run(ctx context.Context) error {
	runner, ctx := utils.NewRunner(ctx, a.logger, a.config.Runner)

	objectStorageClient, err := s3client.NewS3Client(ctx, a.config.S3ClientConfig)
	if err != nil {
		return fmt.Errorf("new s3 client: %w", err)
	}

	dbStorage, err := postgres.NewStorage(a.config.Storage, a.logger)
	if err != nil {
		return fmt.Errorf("new storage: %w", err)
	}
	runner.RunExternal(dbStorage.Shutdown)

	diagramService := diagram.NewService(a.logger, dbStorage, objectStorageClient)

	userService := user.NewService(a.logger)

	httpServer, err := newChartDBServer(ctx, a.logger, a.config.HTTPServer, userService, diagramService)

	runner.RunGraceContext(httpServer.Run, httpServer.Shutdown)

	return runner.Wait()
}

func newChartDBServer(
	ctx context.Context,
	logger *slog.Logger,
	config xhttp.HTTPServerConfig,
	userService user.Service,
	diagramService diagram.Service,
) (*xhttp.HTTPServer, error) {
	chartDBHandler := runtime.NewServeMux(
		runtime.WithErrorHandler(func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			if err := xerrors.HTTPErrorHandler(w, err); err != nil {
				ctxlog.Error(ctx, logger, "http error handler", slog.Any("error", err))
				return
			}

			ctxlog.Info(ctx, logger, "error handled", slog.Any("error", err))
		}),
	)

	err := chartdbapi.RegisterDiagramServiceHandlerServer(
		ctx,
		chartDBHandler,
		&handler.DiagramHandler{
			Logger:         logger,
			DiagramService: diagramService,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("register diagram service handler server: %w", err)
	}

	httpServer, err := xhttp.NewHTTPServer(
		config,
		logger,
		[]func(http.Handler) http.Handler{
			middleware.HTTPAuthMiddleware(logger, userService),
		},
		map[string]http.Handler{
			"/chartdb/v1/diagrams/{id}": chartDBHandler,
			"/chartdb/v1/diagrams":      chartDBHandler,
		})
	if err != nil {
		return nil, fmt.Errorf("new http server: %w", err)
	}

	return httpServer, nil
}
