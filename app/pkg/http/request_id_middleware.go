package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	requestIDHeader = "x-request-id"
	requestIDLogKey = "request_id"
	requestIDLen    = 8
)

type requestIDKey struct{}

func SetRequestID(ctx context.Context, id string) context.Context {
	ctx = ctxlog.WithFields(ctx, slog.String(requestIDLogKey, id))
	ctx = context.WithValue(ctx, requestIDKey{}, id)
	return ctx
}

func GetRequestID(ctx context.Context) string {
	value := ctx.Value(requestIDKey{})
	if requestID, ok := value.(string); ok {
		return requestID
	}

	return ""
}

func randomHexString(bytesLengths int) string {
	buf := make([]byte, bytesLengths)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func GenerateRequestID(ctx context.Context) string {
	requestID := randomHexString(requestIDLen)
	return requestID
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = GenerateRequestID(ctx)
		}
		ctx = SetRequestID(ctx, requestID)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		ww.Header().Add(requestIDHeader, requestID)

		next.ServeHTTP(ww, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
