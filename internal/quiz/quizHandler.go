package quiz

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
	Logger *slog.Logger
	Quiz   *Quiz
}

func New(logger *slog.Logger) (*QuizHandler, error) {

	quiz, err := newQuiz(logger)
	if err != nil {
		logger.Error("failed create quiz", slog.Any("err", err))

		return nil, err
	}

	return &QuizHandler{Logger: logger, Quiz: quiz}, nil
}

func (quizHand *QuizHandler) HandleGetAvailibleNotes(w http.ResponseWriter, _ *http.Request) {

	data, err := json.Marshal(listNotes)
	if err != nil {
		quizHand.Logger.Error("error encoding data", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding data", http.StatusInternalServerError)

		return
	}

	w.Write(data)

}

func (quizHand *QuizHandler) HandleGetNextNote(w http.ResponseWriter, _ *http.Request) {

	NoteForSrv := quizHand.Quiz.getRandomNote()

	data, err := json.Marshal(NoteForSrv)
	if err != nil {
		quizHand.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quizHand.Logger.Info("Отправлен ответ: \n", "data", data)

	w.Write(data)

}

func (quizHand *QuizHandler) HandlePostConfirm(w http.ResponseWriter, r *http.Request) {

	var confirmRequest confirmRequest

	err := json.NewDecoder(r.Body).Decode(&confirmRequest)
	if err != nil {
		quizHand.Logger.Error("error decoding request", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	quizHand.Logger.Info("got input\n", "confirmRequest", confirmRequest)

	confRes := quizHand.Quiz.processConfirmation(&confirmRequest)

	data, err := json.Marshal(confRes)
	if err != nil {
		quizHand.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quizHand.Logger.Info("response\n", "data", data)

	w.Write(data)

}
