package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/JesterForAll/GoNote/internal"
)

func main() {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(slogger)

	serv := internal.NewServer(slogger)

	err := http.ListenAndServe(":9090", serv)

	if err != nil {
		slog.Error("error while serving server", "error", fmt.Errorf("listen and serve: %w", err))
	}
}
