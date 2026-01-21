package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/emiliorevv/api-gateway/internal/limiter"
	"github.com/emiliorevv/api-gateway/internal/mock"
	"github.com/emiliorevv/api-gateway/internal/proxy"
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

	rateLimiter := limiter.NewRateLimiter(rdb, 10, 1.0)
	log.Println("Rate limiter initialized")



	backendURL := mock.Run()
	log.Printf("Backend URL: %s", backendURL)

	proxyHandler, err := proxy.NewReverseProxy(backendURL)
	if err != nil {
		log.Fatal("proxy couldn't be created: ", err)
	}


	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}

		allow, err := rateLimiter.Allow(r.Context(), ip)

		if err != nil{
			log.Printf("Error on rate limiter: %v", err)
			proxyHandler.ServeHTTP(w, r)
			return
		}

		if !allow{
			http.Error(w, "Many petitions", http.StatusTooManyRequests)
			return
		}

		proxyHandler.ServeHTTP(w,r)
	})

	http.Handle("/", finalHandler)

	const port = ":8080"
	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
