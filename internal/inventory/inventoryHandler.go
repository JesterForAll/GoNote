package inventory

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/JesterForAll/gonote/internal/balance"
	"github.com/JesterForAll/gonote/internal/utils"
)

type InventoryHandler struct {
	logger    *slog.Logger
	Inventory *Inventory
}

type responceNumOfSafeFails struct {
	NumOfSafeFails int `json:"NumOfSafeFails"`
}

type helpRequest struct {
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

type helpOctaveResp struct {
	Ok     bool   `json:"ok"`
	Octave string `json:"octave"`
}

type helpNoteResp struct {
	Ok  bool   `json:"ok"`
	Pos string `json:"pos"`
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
	userId, err := utils.GetUserIDFromContext(r.Context(), inventoryHand.logger)
	if err != nil {
		inventoryHand.logger.Error("user_id is missing or invalid")
		http.Error(w, "Missing user_id", http.StatusBadRequest)

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

	_, err = w.Write(data)
	if err != nil {
		inventoryHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}

func (inventoryHand *InventoryHandler) HandlePostUpdateNumOfSafeFails(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.GetUserIDFromContext(r.Context(), inventoryHand.logger)
	if err != nil {
		inventoryHand.logger.Error("user_id is missing or invalid")
		http.Error(w, "Missing user_id", http.StatusBadRequest)

		return
	}

	NumOfSafeFails, err := inventoryHand.Inventory.UpdateCurrentNumOfSafeFailsWithTx(r.Context(), userId, true)
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

	_, err = w.Write(data)
	if err != nil {
		inventoryHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}

func (inventoryHand *InventoryHandler) HandlePostHelpWithOctave(w http.ResponseWriter, r *http.Request) {
	var req helpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		inventoryHand.logger.Error("error decoding request body", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	userID, err := utils.GetUserIDFromContext(r.Context(), inventoryHand.logger)
	if err != nil {
		inventoryHand.logger.Error("user_id is missing or invalid")
		http.Error(w, "Missing user_id", http.StatusBadRequest)

		return
	}

	helpOctaveNumber, err := inventoryHand.Inventory.GetHelpWithOctaveNumber(userID, req.CurrentOctave)
	if err != nil {
		inventoryHand.logger.Error("error getting help with octave", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	responseData := helpOctaveResp{Octave: helpOctaveNumber.Octave, Ok: helpOctaveNumber.Ok}

	data, err := json.Marshal(responseData)
	if err != nil {
		inventoryHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)
	if err != nil {
		inventoryHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}

func (inventoryHand *InventoryHandler) HandlePostHelpWithNote(w http.ResponseWriter, r *http.Request) {
	var req helpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		inventoryHand.logger.Error("error decoding request body", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	userID, err := utils.GetUserIDFromContext(r.Context(), inventoryHand.logger)
	if err != nil {
		inventoryHand.logger.Error("user_id is missing or invalid")
		http.Error(w, "Missing user_id", http.StatusBadRequest)

		return
	}

	helpNotePos, err := inventoryHand.Inventory.GetHelpWithNotePosition(userID, req.CurrentNote)
	if err != nil {
		inventoryHand.logger.Error("error getting help with note", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	responseData := helpNoteResp{Pos: helpNotePos.Pos, Ok: helpNotePos.Ok}

	data, err := json.Marshal(responseData)
	if err != nil {
		inventoryHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)
	if err != nil {
		inventoryHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}
