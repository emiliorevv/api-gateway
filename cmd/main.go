package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/emiliorevv/api-gateway/internal/limiter"
	"github.com/emiliorevv/api-gateway/internal/mock"
	"github.com/emiliorevv/api-gateway/internal/proxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type clientConfig struct {
	Limit int
	Rate float64
}

var clientsInDB = map[string]clientConfig{
	"free-membership-token": {Limit: 5, Rate: 0.5},
	"paid-membership-token": {Limit: 10, Rate: 5.0},
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api-gateway-requestsTotal",
		Help: "Requests processed by the Gateway",
	}, []string {"status"})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	port:= os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}


	rdb, err := limiter.NewRedisClient(redisAddr)
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

			httpRequestsTotal.WithLabelValues("blocked-requests").Inc()

			http.Error(w, "Many petitions", http.StatusTooManyRequests)
			return
		}

		httpRequestsTotal.WithLabelValues("accepted-requests").Inc()
		log.Printf("Allowed rate limiter for api key: %s", apiKey)
		proxyHandler.ServeHTTP(w,r)
	})

	http.Handle("/", finalHandler)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
