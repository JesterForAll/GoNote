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

	mServ := internal.NewServer(slogger)

	err := http.ListenAndServe(":9090", mServ.Serv)

	if err != nil {
		slogger.Error("error while serving server", "error", fmt.Errorf("listen and serve: %w", err))
	}
}
