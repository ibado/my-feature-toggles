package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"myfeaturetoggles.com/toggles/toggles"
	"myfeaturetoggles.com/toggles/util"

	bcrypt "golang.org/x/crypto/bcrypt"
)

var ctx = context.Background()

type SignUpBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	JWT string `json:"jwt"`
}

type sighUpHandler struct {
	repo UserRepository
}

type authHandler struct {
	repo   UserRepository
	logger *log.Logger
}

func NewSignUpHandler(ctx context.Context, logger log.Logger, repo UserRepository) http.Handler {
	return sighUpHandler{repo}
}

func NewAuthUpHandler(ctx context.Context, logger log.Logger, repo UserRepository) http.Handler {
	return authHandler{repo, &logger}
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
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer req.Body.Close()

	var userRequest User
	json.NewDecoder(req.Body).Decode(&userRequest)

	if err := validatePass(userRequest.PasswordHash); err != nil {
		util.ErrorResponse(err, w)
		return
	}

	user, err := h.repo.Get(ctx, userRequest.Email)
	if err != nil {
		util.ErrorResponse(err, w)
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
	Email string
}

func generateJWT(user User) string {
	header := jwtHeader{"HS256"}
	payload := jwtPayload{user.Email}

	headerJson, _ := json.Marshal(header)
	payloadJson, _ := json.Marshal(payload)

	encondedHeader := base64.RawURLEncoding.EncodeToString(headerJson)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJson)

	token := encondedHeader + "." + encodedPayload
	signature := sign(token)

	return token + "." + signature
}

func sign(target string) string {
	hash := hmac.New(sha256.New, []byte(os.Getenv("PRIVATE_KEY")))
	hash.Write([]byte(target))
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

func validatePass(passwordHash string) error {
	return nil
}
