package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/gonote/internal/quiz"
)

type MainServ struct {
	Serv     *http.ServeMux
	Logger   *slog.Logger
	QuizHand *quiz.QuizHandler
}

func New(logger *slog.Logger) *MainServ {
	serv := http.NewServeMux()
	quizHand, err := quiz.New(logger)

	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
	}

	mServ := &MainServ{
		Serv:     serv,
		Logger:   logger,
		QuizHand: quizHand,
	}

	serv.HandleFunc("GET /", mServ.handleGetMain)
	serv.HandleFunc("GET /api/availibleNotes", mServ.QuizHand.HandleGetAvailibleNotes)
	serv.HandleFunc("POST /api/confirm", mServ.QuizHand.HandlePostConfirm)
	serv.HandleFunc("GET /api/new-note", mServ.QuizHand.HandleGetNextNote)
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
