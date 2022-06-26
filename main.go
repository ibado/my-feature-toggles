package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"myfeaturetoggles.com/toggles/toggles"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client = nil
var logger = log.Default()

func health(w http.ResponseWriter, req *http.Request) {
	jsonResponse(map[string]string{"status": "healthy"}, http.StatusOK, w)
}

func errorResponse(err error, w http.ResponseWriter) {
	logger.Println("Error: " + err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}

func jsonResponse(response any, statusCode int, w http.ResponseWriter) {
	json, err := json.Marshal(response)
	if err != nil {
		errorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(json)
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
	if redisClient == nil {
		panic("Fails to connect with Redis")
	}

	repo := toggles.NewRepo(redisClient)
	handleToggles := toggles.NewHandler(ctx, repo, *logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health)
	mux.Handle("/toggles", handleToggles)
	mux.Handle("/toggles/", handleToggles)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
