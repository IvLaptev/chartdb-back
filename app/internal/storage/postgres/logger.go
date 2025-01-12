package postgres

import (
	"context"
	"log/slog"

	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	"github.com/jackc/pgx/v4"
)

const (
	sqlDataField         = "sql"
	postgresPrimaryCheck = "SELECT NOT pg_is_in_recovery()"
)

type Logger struct {
	logger *slog.Logger
}

func newLogger(logger *slog.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]any) {
	var fields []slog.Attr
	if value, ok := data["err"]; ok {
		if err, ok := value.(error); ok {
			if err != nil {
				fields = append(fields, slog.Any("error", err))
			}
			delete(data, "err")
		}
	}
	if len(data) != 0 {
		fields = append(fields, slog.Any("data", data))
	}

	if data[sqlDataField] == postgresPrimaryCheck {
		level = pgx.LogLevelDebug
	}

	switch level {
	case pgx.LogLevelTrace, pgx.LogLevelDebug:
		ctxlog.Debug(ctx, l.logger, msg, fields...)
	case pgx.LogLevelInfo:
		ctxlog.Info(ctx, l.logger, msg, fields...)
	case pgx.LogLevelWarn:
		ctxlog.Warn(ctx, l.logger, msg, fields...)
	case pgx.LogLevelError:
		ctxlog.Error(ctx, l.logger, msg, fields...)
	default:
		ctxlog.Error(ctx, l.logger, msg, append(fields, slog.String("PGX_LOG_LEVEL", level.String()))...)
	}
}
