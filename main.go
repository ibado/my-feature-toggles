package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"myfeaturetoggles.com/toggles/auth"
	"myfeaturetoggles.com/toggles/toggles"
	"myfeaturetoggles.com/toggles/util"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client = nil
var logger = log.Default()

func health(w http.ResponseWriter, req *http.Request) {
	util.JsonResponse(map[string]string{"status": "healthy"}, http.StatusOK, w)
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
	userRepo := auth.NewUserRepo(redisClient)
	handleToggles := toggles.NewHandler(ctx, repo, *logger)
	handleSignUp := auth.NewSignUpHandler(ctx, *logger, userRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health)
	mux.Handle("/toggles", handleToggles)
	mux.Handle("/toggles/", handleToggles)
	mux.Handle("/signup", handleSignUp)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
