package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

type RecoveryConfig struct {
	Enabled bool `yaml:"enabled"`
}

func RecoveryMiddleware(logger *slog.Logger, cfg RecoveryConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if cfg.Enabled {
						// Log panic details
						logger.Error("recovered from panic",
							slog.Any("error", err),
							slog.String("stack", string(debug.Stack())),
						)

						// Return 500 error
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("internal server error"))
					} else {
						// Re-throw panic if recovery is disabled
						panic(err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
