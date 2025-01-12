package ctxlog

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// ContextFields returns log.Fields bound with ctx.
// If no fields are bound, it returns nil.
func ContextFields(ctx context.Context) []slog.Attr {
	fs, _ := ctx.Value(ctxKey{}).([]slog.Attr)
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

func mergeFields(a, b []slog.Attr) []any {
	// NOTE: just append() here is unsafe. If a caller passed slice of fields
	// followed by ... with capacity greater than length, then simultaneous
	// logging will lead to a data race condition.
	//
	// See https://golang.org/ref/spec#Passing_arguments_to_..._parameters
	c := make([]slog.Attr, len(a)+len(b))
	n := copy(c, a)
	copy(c[n:], b)

	res := make([]any, len(c))
	for i := range res {
		res[i] = c[i]
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
