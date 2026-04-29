package balance

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/JesterForAll/gonote/internal/utils"
)

type BalanceHandler struct {
	logger  *slog.Logger
	Balance *Balance
}

type responceBalance struct {
	Balance int `json:"balance"`
}

func New(logger *slog.Logger) (*BalanceHandler, error) {
	balance, err := newBalance(logger)
	if err != nil {
		logger.Error("failed create balance", slog.Any("err", err))

		return nil, err
	}

	return &BalanceHandler{logger: logger, Balance: balance}, nil
}

func (balanceHand *BalanceHandler) HandleGetCurrentBalance(w http.ResponseWriter, r *http.Request) {

	userId, err := utils.GetUserIDFromContext(r.Context(), balanceHand.logger)
	if err != nil {
		balanceHand.logger.Error("user_id is missing or invalid")
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	balance := balanceHand.Balance.GetCurrentBalance(userId)

	responseData := responceBalance{Balance: balance}

	data, err := json.Marshal(responseData)
	if err != nil {
		balanceHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(data)
}
