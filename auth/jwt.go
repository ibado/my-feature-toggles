package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetUserId(req *http.Request) (int64, error) {
	jwt := req.Header.Get("Authorization")
	if jwt == "" {
		return -1, errors.New("No authorization header available")
	}
	split := strings.Split(jwt, ".")
	if len(split) != 3 {
		return -1, errors.New("Invalid JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(split[1])
	if err != nil {
		return -1, err
	}

	var payloadJson jwtPayload
	err = json.Unmarshal(payload, &payloadJson)
	if err != nil {
		return -1, err
	}

	return payloadJson.UserId, nil
}

func generateJWT(user User) string {
	header := jwtHeader{"HS256"}
	payload := jwtPayload{user.Id, time.Now().Unix()}

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
