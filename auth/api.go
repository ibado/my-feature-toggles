package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"myfeaturetoggles.com/toggles/util"

	bcrypt "golang.org/x/crypto/bcrypt"
)

const EXPIRATION_TIME_SECONDS int64 = 2 * 60 * 60 // 2 hs

var ctx = context.Background()

type signUpBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	JWT string `json:"jwt"`
}

type signUpHandler struct {
	repo UserRepository
}

type authHandler struct {
	repo   UserRepository
	logger *log.Logger
}

func NewSignUpHandler(ctx context.Context, logger log.Logger, repo UserRepository) http.Handler {
	return signUpHandler{repo}
}

func NewAuthUpHandler(ctx context.Context, logger log.Logger, repo UserRepository) http.Handler {
	return authHandler{repo, &logger}
}

func (h signUpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	var body signUpBody
	json.NewDecoder(req.Body).Decode(&body)

	if body.Email == "" || body.Password == "" {
		util.JsonError("Both email & password are required", http.StatusBadRequest, w)
		return
	}

	hash, err := hashPass(body.Password)
	if err != nil {
		util.ErrorResponse(err, w)
	}

	_, err = h.repo.Create(ctx, body.Email, hash)
	if err != nil {
		util.ErrorResponse(err, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h authHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer req.Body.Close()

	var userRequest authBody
	json.NewDecoder(req.Body).Decode(&userRequest)
	if userRequest.Email == "" || userRequest.Password == "" {
		util.JsonError("Both email and password are required", http.StatusBadRequest, w)
	}

	user, err := h.repo.Get(ctx, userRequest.Email)
	if err != nil {
		util.ErrorResponse(err, w)
		return
	}
	isValid := validatePass(userRequest.Password, user.PasswordHash)
	if !isValid {
		util.JsonError("Invalid password", 401, w)
		return
	}
	h.logger.Printf("user email: %s, user pass: %s", user.Email, user.PasswordHash)

	token := generateJWT(user)
	util.JsonResponse(AuthResponse{token}, http.StatusOK, w)
}

type jwtHeader struct {
	Algorithm string
}

type jwtPayload struct {
	UserId int64
	Iat    int64
}

func hashPass(password string) (string, error) {
	r, e := bcrypt.GenerateFromPassword([]byte(password), 0)
	if e != nil {
		return "", e
	}
	return string(r), nil
}

func validatePass(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
