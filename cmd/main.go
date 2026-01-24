package main

import (
	"log"
	"net/http"

	"github.com/emiliorevv/api-gateway/internal/limiter"
	"github.com/emiliorevv/api-gateway/internal/mock"
	"github.com/emiliorevv/api-gateway/internal/proxy"
)

type clientConfig struct {
	Limit int
	Rate float64
}

var clientsInDB = map[string]clientConfig{
	"free-membership-token": {Limit: 5, Rate: 0.5},
	"paid-membership-token": {Limit: 10, Rate: 5.0},
}

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

	rateLimiter := limiter.NewRateLimiter(rdb)
	log.Println("Rate limiter initialized")



	backendURL := mock.Run()
	log.Printf("Backend URL: %s", backendURL)

	proxyHandler, err := proxy.NewReverseProxy(backendURL)
	if err != nil {
		log.Fatal("proxy couldn't be created: ", err)
	}


	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")

		config, exists := clientsInDB[apiKey]
		if !exists {
			log.Println("Client not found in database", http.StatusUnauthorized)
			return
		}

		allow, err := rateLimiter.Allow(r.Context(), apiKey, config.Limit, config.Rate)

		if err != nil{
			log.Printf("Error on rate limiter: %v", err)
			proxyHandler.ServeHTTP(w, r)
			return
		}

		if !allow{
			http.Error(w, "Many petitions", http.StatusTooManyRequests)
			return
		}
		log.Printf("Allowed rate limiter for api key: %s", apiKey)
		proxyHandler.ServeHTTP(w,r)
	})

	http.Handle("/", finalHandler)

	const port = ":8080"
	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
