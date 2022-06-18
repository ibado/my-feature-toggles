package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client = nil
var logger = log.Default()

func getToggles() (string, error) {
	toggles, err := redisClient.HGetAll(ctx, "feature-toggles").Result()

	if err != nil {
		logger.Fatal("Error trying to get feature-toggles: " + err.Error())
		return "", err
	}

	jsonMap, err := json.Marshal(toggles)

	if err != nil {
		logger.Fatal("Error marshaling feature-toggles: " + err.Error())
		return "", err
	}

	return string(jsonMap), nil
}

type Body struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

func createToogle(body io.ReadCloser) (error, int) {
	defer body.Close()
	var b Body
	err := json.NewDecoder(body).Decode(&b)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if b.Id == "" || b.Value == "" {
		return errors.New("Both 'id' and 'value' are required"), http.StatusBadRequest
	}
	_, err2 := redisClient.HSet(ctx, "feature-toggles", b.Id, b.Value).Result()
	return err2, http.StatusInternalServerError
}

func toggles(w http.ResponseWriter, req *http.Request) {
	if redisClient == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Redis cliente is not ready!")
		return
	}

	switch req.Method {
	case "GET":
		toggles, err := getToggles()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, toggles)
	case "PUT":
		err, statusCode := createToogle(req.Body)
		if err != nil {
			w.WriteHeader(statusCode)
			io.WriteString(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "Toggle created successfuly")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
	}
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

func health(w http.ResponseWriter, req *http.Request) {
	status, _ := json.Marshal(map[string]string{"status": "healthy"})
	w.Write(status)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	redisClient = createRedisClient()

	http.HandleFunc("/health", health)
	http.HandleFunc("/toggles", toggles)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
