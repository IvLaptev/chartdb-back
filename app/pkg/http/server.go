package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/IvLaptev/chartdb-back/pkg/metrics"
	xmiddleware "github.com/IvLaptev/chartdb-back/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type HTTPServerConfig struct {
	Port      uint64
	TLSConfig HTTPTLSConfig
	CORS      *CORSConfig
	Recovery  xmiddleware.RecoveryConfig `yaml:"recovery"`
}

type HTTPTLSConfig struct {
	Enabled  bool
	CertPath string
	KeyPath  string
}

type CORSConfig struct {
	Origins string `yaml:"origins"`
}

type HTTPServer struct {
	server    *http.Server
	tlsConfig HTTPTLSConfig
}

func (s *HTTPServer) Run() error {
	if s.tlsConfig.Enabled {
		return s.server.ListenAndServeTLS(s.tlsConfig.CertPath, s.tlsConfig.KeyPath)
	} else {
		return s.server.ListenAndServe()
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewHTTPServer(
	cfg HTTPServerConfig,
	logger *slog.Logger,
	middleware []func(http.Handler) http.Handler,
	handlers map[string]http.Handler,
) (*HTTPServer, error) {
	mux := chi.NewMux()

	// Order is important - recovery first, then metrics, then other middleware
	mux.Use(xmiddleware.RecoveryMiddleware(logger, cfg.Recovery))
	mux.Use(metrics.MetricsMiddleware)
	mux.Use(RequestIDMiddleware)
	for _, m := range middleware {
		mux.Use(m)
	}

	if cfg.CORS != nil {
		origins := strings.Split(cfg.CORS.Origins, ",")
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}))
	}

	mux.Handle("/metrics", metrics.Handler())

	for pattern, handler := range handlers {
		mux.Mount(pattern, handler)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return &HTTPServer{
		server:    server,
		tlsConfig: cfg.TLSConfig,
	}, nil
}
