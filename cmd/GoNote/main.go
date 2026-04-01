package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/JesterForAll/GoNote/internal"
)

func main() {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(slogger)

	mServ := internal.NewServer(slogger)

	config, err := ParseConfig()

	if err != nil {
		slogger.Error("error while serving server", "error", fmt.Errorf("listen and serve: %w", err))
	}

	port := config.Port

	slogger.Info("нужно набрать очков для победы", "ScoreToWin", config.ScoreToWin)
	slogger.Info("текущая сложность", "Difficulty", config.Difficulty)

	err = http.ListenAndServe(":"+strconv.Itoa(port), mServ.Serv)

	if err != nil {
		slogger.Error("error while serving server", "error", fmt.Errorf("listen and serve: %w", err))
	}
}
