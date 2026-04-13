package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/gonote/internal/server"
)

func main() {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(slogger)

	mServ := server.New(slogger)

	config, err := ParseConfig()
	if err != nil && errors.Is(err, os.ErrNotExist) {
		slogger.Info("config is not present, using default values for config", "config", *config)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slogger.Error("error parsing config", "error", err)
		os.Exit(1)
	}

	port := config.Port

	slogger.Info("нужно набрать очков для победы", slog.Int("ScoreToWin", config.ScoreToWin))
	slogger.Info("текущая сложность", "Difficulty", slog.Any("Difficulty", config.Difficulty))

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mServ.Serv)
	if err != nil {
		slogger.Error("error while serving server", "error", slog.Any("listen and serve: %w", err))
	}
}
