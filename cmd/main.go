package main

import (
	"GoNote/internal"
	"log"
	"net/http"
)

func main() {
	serv := internal.CreateServer()

	log.Fatal(http.ListenAndServe(":9090", serv))
}
