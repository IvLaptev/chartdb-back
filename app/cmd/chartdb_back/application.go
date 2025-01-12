package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IvLaptev/chartdb-back/internal/handler"
	"github.com/IvLaptev/chartdb-back/internal/service/diagram"
	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
	"github.com/IvLaptev/chartdb-back/internal/utils"
	xhttp "github.com/IvLaptev/chartdb-back/pkg/http"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
	"github.com/go-chi/chi/v5"
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

	httpServer, err := xhttp.NewHTTPServer(a.config.HTTPServer, map[string]chi.Router{
		"/api/diagrams": (&handler.Diagram{
			DiagramService: diagramService,
		}).Router(),
	})
	if err != nil {
		return fmt.Errorf("new http server: %w", err)
	}

	runner.RunGraceContext(httpServer.Run, httpServer.Shutdown)

	return runner.Wait()
}
