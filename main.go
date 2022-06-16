package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client = nil

func hello(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello from myFeatureToggles ;)\n")
}

func toggles(w http.ResponseWriter, req *http.Request) {
	if redisClient == nil {
		w.WriteHeader(500)
		io.WriteString(w, "Redis cliente is not ready!")
	}

	toggles, err := redisClient.HGetAll(ctx, "feature-toggles").Result()

	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "Error trying to get feature-toggles: "+err.Error())
	}

	jsonMap, err := json.Marshal(toggles)

	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "Error marshaling feature-toggles: "+err.Error())
	}

	io.WriteString(w, string(jsonMap))
}

func createRedisClient() *redis.Client {
	url := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")
	return redis.NewClient(&redis.Options{
		Addr:     url,
		Password: pass,
		DB:       0,
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	redisClient = createRedisClient()

	logger := log.Default()
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/toggles", toggles)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
