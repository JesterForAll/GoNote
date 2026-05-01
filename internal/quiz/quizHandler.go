package quiz

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/JesterForAll/gonote/internal/balance"
	"github.com/JesterForAll/gonote/internal/inventory"
)

type availibleNotes struct {
	Octaves []string `json:"octaves"`
	Notes   []string `json:"notes"`
}

type confirmResponse struct {
	Correct     bool    `json:"correct"`
	CorrectNote string  `json:"correctNote"`
	Accuracy    float32 `json:"accuracy"`
}

type confirmRequest struct {
	Note          string `json:"note"`
	Octave        string `json:"octave"`
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

var listNotes = availibleNotes{
	Octaves: []string{"-4", "-3", "-2", "-1", "1", "2", "3", "4", "5"},
	Notes:   []string{"до, C", "до#, C#", "ре, D", "ре#, D#", "ми, E", "фа, F", "фа#, F#", "соль, G", "соль#, G#", "ля, A", "ля#, A#", "си, B"},
}

type QuizHandler struct {
	logger *slog.Logger
	Quiz   *Quiz
}

type noteResponce struct {
	Note     string `json:"note"`
	Octave   string `json:"octave"`
	AudioURL string `json:"audioUrl"`
}

func New(logger *slog.Logger, balance *balance.Balance, inv *inventory.Inventory) (*QuizHandler, error) {
	quiz, err := newQuiz(logger, balance, inv)
	if err != nil {
		logger.Error("failed create quiz", slog.Any("err", err))

		return nil, err
	}

	return &QuizHandler{logger: logger, Quiz: quiz}, nil
}

func (quizHand *QuizHandler) HandleGetAvailibleNotes(w http.ResponseWriter, _ *http.Request) {
	data, err := json.Marshal(listNotes)
	if err != nil {
		quizHand.logger.Error("error encoding data", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding data", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)
	if err != nil {
		quizHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}

func (quizHand *QuizHandler) HandleGetNextNote(w http.ResponseWriter, _ *http.Request) {
	note := quizHand.Quiz.getRandomNote()

	noteForSrv := noteResponce{Note: note.Note, Octave: note.Octave, AudioURL: note.AudioURL}

	data, err := json.Marshal(noteForSrv)
	if err != nil {
		quizHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quizHand.logger.Info("Отправлен ответ: \n", "data", data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)
	if err != nil {
		quizHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}

func (quizHand *QuizHandler) HandlePostConfirm(w http.ResponseWriter, r *http.Request) {
	var confirmRequest confirmRequest

	err := json.NewDecoder(r.Body).Decode(&confirmRequest)
	if err != nil {
		quizHand.logger.Error("error decoding request", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	quizHand.logger.Info("got input\n", "confirm_request", confirmRequest)

	confirm, err := quizHand.Quiz.processConfirmation(r.Context(), &confirmRequest)
	if err != nil {
		quizHand.logger.Error("error saving to database", slog.Any("err", err))
		http.Error(w, "internal server error while saving to database", http.StatusInternalServerError)

		return
	}

	confRes := confirmResponse{Correct: confirm.Correct, CorrectNote: confirm.CorrectNote, Accuracy: confirm.Accuracy}

	data, err := json.Marshal(confRes)
	if err != nil {
		quizHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quizHand.logger.Info("response\n", "data", data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)
	if err != nil {
		quizHand.logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}
}
