package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type HTTPServerConfig struct {
	Port      uint64
	TLSConfig HTTPTLSConfig
	CORS      *CORSConfig
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
	routers map[string]chi.Router,
) (*HTTPServer, error) {
	mux := chi.NewMux()

	// mux.Use(metrics.NewHTTPMetricsMiddleware(registry))
	// mux.Use(tracer.NewHTTPTraceMiddleware(tracer.WithExcludedPath("/metrics")))

	if cfg.CORS != nil {
		origins := strings.Split(cfg.CORS.Origins, ",")
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}))
	}

	for pattern, router := range routers {
		mux.Mount(pattern, router)
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
