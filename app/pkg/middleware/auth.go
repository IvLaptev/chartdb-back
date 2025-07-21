package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
)

const (
	XUserIDHeader = "x-user-id"
)

type authService interface {
	Authenticate(ctx context.Context, token string) (context.Context, error)
}

func HTTPAuthMiddleware(logger *slog.Logger, authService authService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get(XUserIDHeader)
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx, err := authService.Authenticate(ctx, authHeader)
			if err != nil {
				if err := xerrors.HTTPErrorHandler(w, err); err != nil {
					ctxlog.Error(ctx, logger, "http error handler", slog.Any("error", err))
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
