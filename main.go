package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello from myFeatureToggles ;)\n")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	logger := log.Default()
	http.HandleFunc("/hello", hello)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
