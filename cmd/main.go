package main

import (
	"github.com/emiliorevv/api-gateway/internal/mock"

	"log"
	
	"net/http"

	"github.com/emiliorevv/api-gateway/internal/proxy"
)

func main() {
	backendURL := mock.Run()
	log.Printf("Backend URL: %s", backendURL)

	proxyHandler, err := proxy.NewReverseProxy(backendURL)
	if err != nil {
		log.Fatal("proxy couldnt be created: ", err)
	}

	http.Handle("/", proxyHandler)

	const port = ":8080"
	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
