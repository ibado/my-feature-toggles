package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func ErrorResponse(err error, w http.ResponseWriter) {
	log.Default().Println("Error: " + err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}

func JsonResponse(response any, statusCode int, w http.ResponseWriter) {
	json, err := json.Marshal(response)
	if err != nil {
		ErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(json)
}
