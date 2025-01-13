package ctxlog

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// ContextFields returns log.Fields bound with ctx.
// If no fields are bound, it returns nil.
func ContextFields(ctx context.Context) []any {
	fs, _ := ctx.Value(ctxKey{}).([]any)

	return fs
}

// WithFields returns a new context that is bound with given fields and based
// on parent ctx.
func WithFields(ctx context.Context, fields ...slog.Attr) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	return context.WithValue(ctx, ctxKey{}, mergeFields(ContextFields(ctx), fields))
}

// Debug logs at Debug log level using fields both from arguments and ones that
// are bound to ctx.
func Debug(ctx context.Context, l *slog.Logger, msg string, fields ...slog.Attr) {
	l.Debug(msg, mergeFields(ContextFields(ctx), fields)...)
}

// Info logs at Info log level using fields both from arguments and ones that
// are bound to ctx.
func Info(ctx context.Context, l *slog.Logger, msg string, fields ...slog.Attr) {
	l.Info(msg, mergeFields(ContextFields(ctx), fields)...)
}

// Warn logs at Warn log level using fields both from arguments and ones that
// are bound to ctx.
func Warn(ctx context.Context, l *slog.Logger, msg string, fields ...slog.Attr) {
	l.Warn(msg, mergeFields(ContextFields(ctx), fields)...)
}

// Error logs at Error log level using fields both from arguments and ones that
// are bound to ctx.
func Error(ctx context.Context, l *slog.Logger, msg string, fields ...slog.Attr) {
	l.Error(msg, mergeFields(ContextFields(ctx), fields)...)
}

// Debugf logs at Debug log level using fields that are bound to ctx.
// The message is formatted using provided arguments.
// func Debugf(ctx context.Context, l *slog.Logger, format string, args ...interface{}) {
// 	msg := fmt.Sprintf(format, args...)
// 	l.Debug(msg, ContextFields(ctx))
// }

// Infof logs at Info log level using fields that are bound to ctx.
// The message is formatted using provided arguments.
// func Infof(ctx context.Context, l *slog.Logger, format string, args ...interface{}) {
// 	msg := fmt.Sprintf(format, args...)
// 	l.Info(msg, ContextFields(ctx))
// }

// Warnf logs at Warn log level using fields that are bound to ctx.
// The message is formatted using provided arguments.
// func Warnf(ctx context.Context, l *slog.Logger, format string, args ...interface{}) {
// 	msg := fmt.Sprintf(format, args...)
// 	l.Warn(msg, ContextFields(ctx))
// }

// Errorf logs at Error log level using fields that are bound to ctx.
// The message is formatted using provided arguments.
// func Errorf(ctx context.Context, l *slog.Logger, format string, args ...interface{}) {
// 	msg := fmt.Sprintf(format, args...)
// 	l.Error(msg, ContextFields(ctx))
// }

func mergeFields(a []any, b []slog.Attr) []any {
	res := make([]any, len(a)+len(b))
	copy(res, a)
	for i := 0; i < len(b); i++ {
		res[len(a)+i] = b[i]
	}
	return res
}

func WriteAt(lvl slog.Level, ctx context.Context, l *slog.Logger, msg string, fields ...slog.Attr) {
	switch lvl {
	case slog.LevelDebug:
		Debug(ctx, l, msg, fields...)
	case slog.LevelInfo:
		Info(ctx, l, msg, fields...)
	case slog.LevelWarn:
		Warn(ctx, l, msg, fields...)
	case slog.LevelError:
		Error(ctx, l, msg, fields...)
	}
}

// func WriteAtf(lvl slog.Level, ctx context.Context, l *slog.Logger, format string, args ...interface{}) {
// 	l = l

// 	switch lvl {
// 	case slog.LevelDebug:
// 		Debugf(ctx, l, format, args...)
// 	case slog.LevelInfo:
// 		Infof(ctx, l, format, args...)
// 	case slog.LevelWarn:
// 		Warnf(ctx, l, format, args...)
// 	case slog.LevelError:
// 		Errorf(ctx, l, format, args...)
// 	}
// }
