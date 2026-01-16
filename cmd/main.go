package main

import (
	"github.com/emiliorevv/api-gateway/internal/mock"

	"log"

	"net/http"

	"github.com/emiliorevv/api-gateway/internal/proxy"

	"github.com/emiliorevv/api-gateway/internal/limiter"
)

func main() {

	rdb, err := limiter.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatal("Redis connection error: ", err)
	}

	defer func() {
		err := rdb.Close()
		if err != nil {
			log.Fatal("Error closing redis connection: ", err)
		}
	}()

	log.Printf("Connection to redis successful")

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
