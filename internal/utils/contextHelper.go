package utils

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/JesterForAll/gonote/internal/contextkey"
)

func GetUserIDFromContext(ctx context.Context, logger *slog.Logger) (int, error) {
	val := ctx.Value(contextkey.GetUserIDKey())
	if val == nil {
		logger.Error("userID not found in context")
		return 0, strconv.ErrSyntax
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case uint:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		logger.Error("unexpected userID type in context", slog.Any("type", v))
		return 0, strconv.ErrSyntax
	}
}
