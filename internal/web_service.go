package internal

import (
	"net/http"
	"os"
)

func getMainHandle() http.Handler {
	data, err := os.ReadFile("../static/index.html")

	if err != nil {
		panic(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})

	return handler
}

func CreateServer() *http.ServeMux {
	serv := http.NewServeMux()

	serv.Handle("/", getMainHandle())
	serv.Handle("/api/confirm", nil)
	serv.Handle("/api/availibleNotes", nil)
	serv.Handle("/api/new-note", nil)
	serv.Handle("/api/prev-note", nil)

	return serv
}
