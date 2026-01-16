package main

import (
	"fmt"

	"github.com/emiliorevv/api-gateway/internal/mock"

	"log"

	"github.com/emiliorevv/api-gateway/internal/mock"

	"net/http"
)

func main() {
	backendURL := mock.Run()
	log.Printf("Backend URL: %s", backendURL)

	const port = ":8080"
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "pong")

	})
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal(err)
	}
}
