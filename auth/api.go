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
	"strings"
	"time"

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

	user := User{body.Email, hash}
	err = h.repo.Create(ctx, user)
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
	Email string
	Iat   int64
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

func generateJWT(user User) string {
	header := jwtHeader{"HS256"}
	payload := jwtPayload{user.Email, time.Now().Unix()}

	headerJson, _ := json.Marshal(header)
	payloadJson, _ := json.Marshal(payload)

	encondedHeader := base64.RawURLEncoding.EncodeToString(headerJson)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJson)

	headerWithPayload := encondedHeader + "." + encodedPayload
	signature := sign(headerWithPayload)

	return headerWithPayload + "." + signature
}

func sign(target string) string {
	hash := hmac.New(sha256.New, []byte(os.Getenv("PRIVATE_KEY")))
	hash.Write([]byte(target))
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

func validateJWT(token string) bool {
	split := strings.Split(token, ".")
	if len(split) != 3 {
		return false
	}

	headerAndPayload := split[0] + "." + split[1]
	signature := split[2]

	if sign(headerAndPayload) != signature {
		return false
	}

	payload, _ := decodeJWTPayload(split[1])
	if payload.Iat+EXPIRATION_TIME_SECONDS < time.Now().Unix() {
		return false
	}

	return true
}

func decodeJWTPayload(payload string) (jwtPayload, error) {
	jsonPayload, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return jwtPayload{}, err
	}

	var jp jwtPayload
	err = json.Unmarshal(jsonPayload, &jp)
	if err != nil {
		return jwtPayload{}, err
	}

	return jp, nil
}
