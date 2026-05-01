package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/gonote/internal/balance"
	"github.com/JesterForAll/gonote/internal/inventory"
	"github.com/JesterForAll/gonote/internal/jwt"
	"github.com/JesterForAll/gonote/internal/login"
	"github.com/JesterForAll/gonote/internal/middleware"
	"github.com/JesterForAll/gonote/internal/quiz"
)

type MainServ struct {
	Serv           *http.ServeMux
	JWTManager     *jwt.Manager
	Logger         *slog.Logger
	QuizHand       *quiz.QuizHandler
	LoginHand      *login.LoginHandler
	InventoryHand  *inventory.InventoryHandler
	BalanceHandler *balance.BalanceHandler
}

func New(logger *slog.Logger) (*MainServ, error) {
	serv := http.NewServeMux()

	jwtManager := jwt.NewManager()

	loginHand, err := login.New(logger, jwtManager)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	balanceHand, err := balance.New(logger)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	invHand, err := inventory.New(logger, balanceHand.Balance)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	quizHand, err := quiz.New(logger, balanceHand.Balance, invHand.Inventory)
	if err != nil {
		logger.Error("internal error", slog.Any("err", err))
		return nil, err
	}

	mServ := &MainServ{
		Serv:           serv,
		Logger:         logger,
		JWTManager:     jwtManager,
		QuizHand:       quizHand,
		LoginHand:      loginHand,
		InventoryHand:  invHand,
		BalanceHandler: balanceHand,
	}

	publicMux := http.NewServeMux()

	publicMux.HandleFunc("GET /", mServ.handleGetRoot)
	publicMux.HandleFunc("GET /api/login/getUsers", mServ.LoginHand.HandleGetUsers)
	publicMux.HandleFunc("POST /api/login/createUser", mServ.LoginHand.HandleCreateUser)
	publicMux.HandleFunc("POST /api/login/login", mServ.LoginHand.HandleLogin)

	authMux := http.NewServeMux()

	authMux.HandleFunc("GET /main", mServ.handleGetMain)
	authMux.HandleFunc("GET /api/availibleNotes", mServ.QuizHand.HandleGetAvailibleNotes)
	authMux.HandleFunc("POST /api/confirm", mServ.QuizHand.HandlePostConfirm)
	authMux.HandleFunc("GET /api/new-note", mServ.QuizHand.HandleGetNextNote)

	authMux.HandleFunc("GET /api/balance", mServ.BalanceHandler.HandleGetCurrentBalance)

	authMux.HandleFunc("GET /api/num-of-safe-fails", mServ.InventoryHand.HandleGetCurrentBalance)
	authMux.HandleFunc("POST /api/update-safe-fails", mServ.InventoryHand.HandlePostUpdateNumOfSafeFails)
	authMux.HandleFunc("POST /api/buy-note-help", mServ.InventoryHand.HandlePostHelpWithNote)
	authMux.HandleFunc("POST /api/buy-octave-help", mServ.InventoryHand.HandlePostHelpWithOctave)

	authHanlder := middleware.NewUserContextMiddleware(jwtManager, authMux, logger)

	serv.Handle("/", publicMux)
	serv.Handle("/api/login/", publicMux)
	serv.Handle("/api/", authHanlder)
	serv.Handle("/main", authHanlder)

	serv.Handle("GET /notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../../assets"))))

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

	_, err = w.Write(data)
	if err != nil {
		mServ.Logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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

	_, err = w.Write(data)
	if err != nil {
		mServ.Logger.Error("error writing data", slog.Any("err", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}
