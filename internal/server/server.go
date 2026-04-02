package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/gonote/internal/quiz"
)

type MainServ struct {
	Serv   *http.ServeMux
	Logger *slog.Logger
	Quiz   *quiz.Quiz
}

func New(logger *slog.Logger) *MainServ {
	serv := http.NewServeMux()
	quiz, err := quiz.New(logger)

	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
	}

	mServ := &MainServ{
		Serv:   serv,
		Logger: logger,
		Quiz:   quiz,
	}

	serv.HandleFunc("GET /", mServ.handleGetMain)
	serv.HandleFunc("GET /api/availibleNotes", mServ.Quiz.HandleGetAvailibleNotes)
	serv.HandleFunc("POST /api/confirm", mServ.Quiz.HandlePostConfirm)
	serv.HandleFunc("GET /api/new-note", mServ.Quiz.HandleGetNextNote)
	// serv.Handle("/api/prev-note", nil)
	serv.Handle("GET /notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../../assets"))))

	return mServ
}

func (mServ *MainServ) handleGetMain(w http.ResponseWriter, _ *http.Request) {

	data, err := os.ReadFile("../../static/index.html")

	if err != nil {
		mServ.Logger.Error("error reading main page", slog.Any("err", err))
		http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

		return
	}

	w.Write(data)

}
