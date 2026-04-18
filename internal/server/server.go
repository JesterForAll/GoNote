package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/gonote/internal/login"
	"github.com/JesterForAll/gonote/internal/quiz"
)

type MainServ struct {
	Serv      *http.ServeMux
	Logger    *slog.Logger
	QuizHand  *quiz.QuizHandler
	LoginHand *login.LoginHandler
}

func New(logger *slog.Logger) (*MainServ, error) {
	serv := http.NewServeMux()

	quizHand, err := quiz.New(logger)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	loginHand, err := login.New(logger)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	mServ := &MainServ{
		Serv:      serv,
		Logger:    logger,
		QuizHand:  quizHand,
		LoginHand: loginHand,
	}

	serv.HandleFunc("GET /", mServ.handleGetRoot)
	serv.HandleFunc("GET /main", mServ.handleGetMain)

	serv.HandleFunc("GET /api/availibleNotes", mServ.QuizHand.HandleGetAvailibleNotes)
	serv.HandleFunc("POST /api/confirm", mServ.QuizHand.HandlePostConfirm)
	serv.HandleFunc("GET /api/new-note", mServ.QuizHand.HandleGetNextNote)
	serv.Handle("GET /notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../../assets"))))

	serv.HandleFunc("GET /api/getUsers", mServ.LoginHand.HandleGetUsers)
	serv.HandleFunc("POST /api/createUser", mServ.LoginHand.HandleCreateUser)
	serv.HandleFunc("POST /api/login", mServ.LoginHand.HandleLogin)

	return mServ, nil
}

func (mServ *MainServ) handleGetRoot(w http.ResponseWriter, _ *http.Request) {

	data, err := os.ReadFile("../../static/login.html")

	if err != nil {
		mServ.Logger.Error("error reading main page", slog.Any("err", err))
		http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write(data)

}

func (mServ *MainServ) handleGetMain(w http.ResponseWriter, _ *http.Request) {

	data, err := os.ReadFile("../../static/index.html")

	if err != nil {
		mServ.Logger.Error("error reading main page", slog.Any("err", err))
		http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write(data)

}
