package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"myfeaturetoggles.com/toggles/auth"
	"myfeaturetoggles.com/toggles/toggles"
	"myfeaturetoggles.com/toggles/util"

	_ "github.com/lib/pq"
)

var ctx = context.Background()
var dbConnection *sql.DB = nil
var logger = log.Default()

func health(w http.ResponseWriter, req *http.Request) {
	util.JsonResponse(map[string]string{"status": "healthy"}, http.StatusOK, w)
}

func createDBConnection() *sql.DB {
	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return db
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbConnection = createDBConnection()
	if dbConnection == nil {
		panic("Fails to connect with Postgres")
	}

	repo := toggles.NewRepo(dbConnection)
	userRepo := auth.NewUserRepo(dbConnection)
	handleToggles := toggles.NewHandler(ctx, repo, *logger)
	handleSignUp := auth.NewSignUpHandler(ctx, *logger, userRepo)
	handleAuth := auth.NewAuthUpHandler(ctx, *logger, userRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health)
	mux.Handle("/toggles", handleToggles)
	mux.Handle("/toggles/", handleToggles)
	mux.Handle("/signup", handleSignUp)
	mux.Handle("/auth", handleAuth)

	logger.Println("running server on port " + port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
