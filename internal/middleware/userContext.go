package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JesterForAll/gonote/internal/contextkey"
	"github.com/JesterForAll/gonote/internal/jwt"
)

type UserContextMiddleware struct {
	jwtManager *jwt.Manager
	next       http.Handler
	logger     *slog.Logger
}

func NewUserContextMiddleware(jwtManager *jwt.Manager, next http.Handler, logger *slog.Logger) *UserContextMiddleware {
	return &UserContextMiddleware{
		jwtManager: jwtManager,
		next:       next,
		logger:     logger,
	}
}

func (m *UserContextMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		m.logger.Error("No Authorization header")
		http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)

		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		m.logger.Error("Invalid authorization header format")
		http.Error(w, "Unauthorized: invalid authorization format", http.StatusUnauthorized)

		return
	}

	tokenString := parts[1]

	userID, err := m.jwtManager.ParseToken(tokenString)
	if err != nil {
		m.logger.Error("Invalid or expired JWT token", slog.Any("err", err))
		http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)

		return
	}

	ctx := context.WithValue(r.Context(), contextkey.GetUserIDKey(), userID)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}
