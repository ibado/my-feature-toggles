package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"myfeaturetoggles.com/toggles/toggles"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client = nil
var logger = log.Default()

type Toggle struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type Ups struct {
	Msg string `json:"error"`
}

func handleToggles(w http.ResponseWriter, req *http.Request) {
	if redisClient == nil {
		errorResponse(errors.New("Redis cliente is not ready!"), w)
		return
	}

	repo := toggles.NewRepo(*redisClient)

	switch req.Method {
	case "GET":
		toggles, err := repo.GetAll(ctx)
		if err != nil {
			errorResponse(err, w)
			return
		}
		jsonResponse(toggles, http.StatusOK, w)
	case "PUT":
		defer req.Body.Close()
		var toggle Toggle
		err := json.NewDecoder(req.Body).Decode(&toggle)
		if err != nil || toggle.Id == "" || toggle.Value == "" {
			res := Ups{"Both 'id' and 'value' are required"}
			jsonResponse(res, http.StatusBadRequest, w)
			return
		}
		err = repo.Add(ctx, toggle.Id, toggle.Value)
		if err != nil {
			errorResponse(err, w)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case "DELETE":
		id := strings.Replace(req.URL.Path, "/toggles/", "", -1)

		if len(id) == 0 || id == req.URL.Path {
			res := Ups{"A valid id is required for removing a toggle"}
			jsonResponse(res, http.StatusBadRequest, w)
			return
		}

		exist, err := repo.Exist(ctx, id)
		if err != nil {
			errorResponse(err, w)
			return
		}
		if !exist {
			msg := fmt.Sprintf("the id '%s' doesn't match with an existing toggle", id)
			jsonResponse(Ups{msg}, http.StatusBadRequest, w)
			return
		}

		err = repo.Remove(ctx, id)
		if err != nil {
			errorResponse(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		jsonResponse(Ups{"Method not allowed"}, http.StatusMethodNotAllowed, w)
	}
}

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

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health)
	mux.HandleFunc("/toggles", handleToggles)
	mux.HandleFunc("/toggles/", handleToggles)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
