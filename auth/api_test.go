package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepo struct {
	User
}

func (fr fakeRepo) Create(context.Context, User) error {
	return nil
}

func (fr fakeRepo) Get(ctx context.Context, email string) (User, error) {
	return fr.User, nil
}

func TestSignUp(t *testing.T) {

	body, _ := json.Marshal(signUpBody{"ibado", "pass1234"})

	handler := NewSignUpHandler(context.Background(), *log.Default(), fakeRepo{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))

	handler.ServeHTTP(recorder, request)

	result := recorder.Result()

	if result.StatusCode != 201 {
		t.Fatal("StatusCode should be 201")
	}
}

func TestSignUpFail(t *testing.T) {

	body, _ := json.Marshal(signUpBody{"", "pass1234"})

	handler := NewSignUpHandler(context.Background(), *log.Default(), fakeRepo{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))

	handler.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != 400 {
		t.Fatal("StatusCode should be 400")
	}

	var ups map[string]string
	json.NewDecoder(result.Body).Decode(&ups)
	if ups["error"] != "Both email & password are required to be not empty" {
		t.Fatal("Wrong response body msg")
	}
}

func validJWT(jwt string) bool {
	return len(jwt) > 20 && strings.Contains(jwt, ".")
}

func TestAuth(t *testing.T) {
	ab := authBody{Email: "test@test.com", Password: "asd123456"}
	passwordHash, err := hashPass(ab.Password)
	user := User{"test@test.com", passwordHash}
	repo := fakeRepo{user}
	authHandler := NewAuthUpHandler(context.Background(), *log.Default(), repo)
	recorder := httptest.NewRecorder()
	body, err := json.Marshal(ab)
	if err != nil {
		t.Fail()
	}
	request := httptest.NewRequest("POST", "/auth", bytes.NewReader(body))

	authHandler.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != 200 {
		t.Fatalf("Status code should be 200 but is %d", result.StatusCode)
	}

	var resultBody AuthResponse
	json.NewDecoder(result.Body).Decode(&resultBody)

	if !validJWT(resultBody.JWT) {
		t.Fatal("invalid JWT")
	}
}

func TestAuthInvalidPass(t *testing.T) {
	ab := authBody{Email: "test@test.com", Password: "invalid password"}
	user := User{"test@test.com", "hash that doesn't match"}
	repo := fakeRepo{user}
	authHandler := NewAuthUpHandler(context.Background(), *log.Default(), repo)
	recorder := httptest.NewRecorder()
	body, err := json.Marshal(ab)
	if err != nil {
		t.Fail()
	}
	request := httptest.NewRequest("POST", "/auth", bytes.NewReader(body))

	authHandler.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != 401 {
		t.Fatalf("Status code should be 401 but is %d", result.StatusCode)
	}
}
