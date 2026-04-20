package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/JesterForAll/gonote/internal/contextkey"
	"github.com/JesterForAll/gonote/internal/session"
)

type UserContextMiddleware struct {
	tokenManager *session.TokenManager
	next         http.Handler
	logger       *slog.Logger
}

func NewUserContextMiddleware(tokenManager *session.TokenManager, next http.Handler, logger *slog.Logger) *UserContextMiddleware {
	return &UserContextMiddleware{
		tokenManager: tokenManager,
		next:         next,
		logger:       logger,
	}
}

func (m *UserContextMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		m.logger.Error("No coockie availible")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) != 2 {
		m.logger.Error("Invalid cookie format")
		http.Error(w, "Invalid cookie format", http.StatusBadRequest)

		return
	}

	currentToken := m.tokenManager.GetToken()
	if parts[1] != currentToken {
		m.logger.Error("Session expired")
		http.Error(w, "Session expired", http.StatusUnauthorized)

		return
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		m.logger.Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)

		return
	}

	ctx := context.WithValue(r.Context(), contextkey.GetUserIDKey(), userID)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}
