package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	redis "github.com/go-redis/redis/v8"
)

const TOGGLES_KEY = "feature-toggles"

var ctx = context.Background()
var redisClient *redis.Client = nil
var logger = log.Default()

func getToggles() (string, error) {
	toggles, err := redisClient.HGetAll(ctx, TOGGLES_KEY).Result()

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
	_, err2 := redisClient.HSet(ctx, TOGGLES_KEY, b.Id, b.Value).Result()
	return err2, http.StatusInternalServerError
}

func removeToggle(id string) error {
	exist, _ := redisClient.HExists(ctx, TOGGLES_KEY, id).Result()
	if !exist {
		msg := fmt.Sprintf("the id '%s' doesn't match with an existing toggle", id)
		return errors.New(msg)
	}
	err := redisClient.HDel(ctx, TOGGLES_KEY, id).Err()
	return err
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
	case "DELETE":
		id := strings.Replace(req.URL.Path, "/toggles/", "", -1)

		if len(id) == 0 || id == req.URL.Path {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "A valid id is required for removing a toggle")
			return
		}

		err := removeToggle(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "Toggle removed successfuly")
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
	http.HandleFunc("/toggles/", toggles)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
