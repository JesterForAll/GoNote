package inventory

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JesterForAll/gonote/internal/balance"
)

type InventoryHandler struct {
	logger    *slog.Logger
	Inventory *Inventory
}

type responceNumOfSafeFails struct {
	NumOfSafeFails int
}

type reqUserID struct {
	UserID string `json:"user_id"`
}

func New(logger *slog.Logger, balance *balance.Balance) (*InventoryHandler, error) {
	inventory, err := newInventory(logger, balance)
	if err != nil {
		logger.Error("failed create balance", slog.Any("err", err))

		return nil, err
	}

	return &InventoryHandler{logger: logger, Inventory: inventory}, nil
}

func (inventoryHand *InventoryHandler) HandleGetCurrentBalance(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("user_id")

	if userIdStr == "" {
		inventoryHand.logger.Error("user_id is missing")
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		inventoryHand.logger.Error("invalid user_id", slog.Any("err", err))
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	NumOfSafeFails := inventoryHand.Inventory.GetCurrentNumOfSafeFails(userId)

	responseData := responceNumOfSafeFails{NumOfSafeFails: NumOfSafeFails}

	data, err := json.Marshal(responseData)
	if err != nil {
		inventoryHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(data)
}

func (inventoryHand *InventoryHandler) HandlePostUpdateNumOfSafeFails(w http.ResponseWriter, r *http.Request) {
	var reqUserID reqUserID

	err := json.NewDecoder(r.Body).Decode(&reqUserID)
	if err != nil {
		inventoryHand.logger.Error("error decoding create user request", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	userId, err := strconv.Atoi(reqUserID.UserID)
	if err != nil {
		inventoryHand.logger.Error("invalid user_id", slog.Any("err", err))
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	NumOfSafeFails, err := inventoryHand.Inventory.UpdateCurrentNumOfSafeFails(userId, true)
	if err != nil {
		inventoryHand.logger.Error("error updating number of safe fails", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseData := responceNumOfSafeFails{NumOfSafeFails: NumOfSafeFails}

	data, err := json.Marshal(responseData)
	if err != nil {
		inventoryHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(data)
}
