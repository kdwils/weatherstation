package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWithContext(t *testing.T) {
	logger := zap.NewExample()
	ctx := context.Background()

	newCtx := WithContext(ctx, logger)
	assert.NotNil(t, newCtx)

	extractedLogger, err := FromContext(newCtx)
	assert.NoError(t, err)
	assert.Equal(t, logger, extractedLogger)
}

func TestFromContext_NoLogger(t *testing.T) {
	ctx := context.Background()
	logger, err := FromContext(ctx)

	assert.Error(t, err)
	assert.Nil(t, logger)
	assert.Equal(t, "no logger found", err.Error())
}

func TestFromContext_WithLogger(t *testing.T) {
	logger := zap.NewExample()
	ctx := WithContext(context.Background(), logger)

	extractedLogger, err := FromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, logger, extractedLogger)
}
