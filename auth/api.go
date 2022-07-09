package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"myfeaturetoggles.com/toggles/toggles"
	"myfeaturetoggles.com/toggles/util"

	bcrypt "golang.org/x/crypto/bcrypt"
)

var ctx = context.Background()

type SignUpBody struct {
	Email    string
	Password string
}

type sighUpHandler struct {
	repo UserRepository
}

type authHandler struct {
	repo UserRepository
}

func NewSignUpHandler(ctx context.Context, logger log.Logger, repo UserRepository) http.Handler {
	return sighUpHandler{repo}
}

func NewAuthUpHandler(ctx context.Context, logger log.Logger) http.Handler {
	return authHandler{}

}

func (h sighUpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	var body SignUpBody
	json.NewDecoder(req.Body).Decode(&body)

	if body.Email == "" || body.Password == "" {
		msg := "Both email & password are required to be not empty"
		util.JsonResponse(toggles.Ups{Msg: msg}, http.StatusBadRequest, w)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 0)
	if err != nil {
		util.ErrorResponse(err, w)
	}

	user := User{body.Email, string(hash)}
	err = h.repo.Create(ctx, user)
	if err != nil {
		util.ErrorResponse(err, w)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h authHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

}
