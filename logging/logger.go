package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type zapLoggerKey struct{}

var (
	loggerKey = zapLoggerKey{}
)

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) (*zap.Logger, error) {
	if v, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return v, nil
	}

	return nil, fmt.Errorf("no logger found")
}
