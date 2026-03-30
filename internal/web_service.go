package internal

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type availibleNotes struct {
	Octaves []string `json:"octaves"`
	Notes   []string `json:"notes"`
}

type note struct {
	Note     string `json:"note"`
	Octave   string `json:"octave"`
	Index    int    `json:"index"`
	AudioUrl string `json:"audioUrl"`
}

type confirmResponse struct {
	Correct     bool   `json:"correct"`
	CorrectNote string `json:"correctNote"`
}

type confirmRequest struct {
	Note          string `json:"note"`
	Octave        string `json:"octave"`
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

func getMainHandle(logger *slog.Logger) http.Handler {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		data, err := os.ReadFile("../static/index.html")

		if err != nil {
			logger.Error("error reading main page", slog.Any("err", err))
			http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

			return
		}

		w.Write(data)
	})

	return handler
}

var ListNotes = availibleNotes{
	Octaves: []string{"-4", "-3", "-2", "-1", "1", "2", "3", "4", "5"},
	Notes:   []string{"до, C", "до#, C#", "ре, D", "ре#, D#", "ми, E", "фа, F", "фа#, F#", "соль, G", "соль#, G#", "ля, A", "ля#, A#", "си, B"},
}

func getAvailibleNotes(logger *slog.Logger) http.Handler {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(ListNotes)

		if err != nil {
			logger.Error("error encoding data", slog.Any("err", err))
			http.Error(w, "Internal server error while encoding data", http.StatusInternalServerError)

			return
		}

		w.Write(data)
	})

	return handler
}

func getNewNote(logger *slog.Logger) http.Handler {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileList, err := os.ReadDir("../assets")

		if err != nil {
			logger.Error("error getting data from disk", slog.Any("err", err))
			http.Error(w, "Internal server error while getting data from disk", http.StatusInternalServerError)

			return
		}

		newR := rand.New(rand.NewSource(time.Now().UnixNano()))
		randNote := fileList[newR.Intn(len(fileList))]

		// note name is in format "1-я октава (1), ре, D.wav"
		var NoteForSrv note

		indxLead := strings.Index(randNote.Name(), "(")
		indxEnd := strings.Index(randNote.Name(), ")")

		if indxLead != -1 && indxEnd != -1 {
			NoteForSrv.Octave = randNote.Name()[indxLead+1 : indxEnd]
		}

		indxDot := strings.Index(randNote.Name(), ".")
		if indxDot != -1 {
			NoteForSrv.Note = strings.TrimSpace(randNote.Name()[indxEnd+2 : indxDot])
		}

		NoteForSrv.AudioUrl = randNote.Name() //"http://localhost:9090/notes/" +

		data, err := json.Marshal(NoteForSrv)

		if err != nil {
			logger.Error("error encoding response", slog.Any("err", err))
			http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

			return
		}

		logger.Info("Отправлен ответ: \n", "data", data)

		w.Write(data)
	})

	return handler
}

func getConfirm(logger *slog.Logger) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var confirmRequest confirmRequest

		err := json.NewDecoder(r.Body).Decode(&confirmRequest)

		if err != nil {
			logger.Error("error decoding request", slog.Any("err", err))
			http.Error(w, "Internal server error while decoding body", http.StatusInternalServerError)

			return
		}

		logger.Info("got input\n", "confirmRequest", confirmRequest)

		var confRes confirmResponse

		if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
			confRes.Correct = true
		} else {
			confRes.Correct = false
			confRes.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
		}

		data, err := json.Marshal(confRes)

		if err != nil {
			logger.Error("error encoding response", slog.Any("err", err))
			http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

			return
		}

		logger.Info("response\n", "data", data)

		w.Write(data)
	})

	return handler
}

func NewServer(logger *slog.Logger) *http.ServeMux {
	serv := http.NewServeMux()

	serv.Handle("GET /", getMainHandle(logger))
	serv.Handle("GET /api/availibleNotes", getAvailibleNotes(logger))
	serv.Handle("POST /api/confirm", getConfirm(logger))
	serv.Handle("GET /api/new-note", getNewNote(logger))
	// serv.Handle("/api/prev-note", nil)
	serv.Handle("GET /notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../assets"))))

	return serv
}
