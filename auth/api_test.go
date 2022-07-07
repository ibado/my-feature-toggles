package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	"myfeaturetoggles.com/toggles/toggles"
)

type fakeRepo struct {
}

func (fr fakeRepo) Create(context.Context, User) error {
	return nil
}

func TestSignUp(t *testing.T) {

	body, _ := json.Marshal(SignUpBody{"ibado", "pass1234"})

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

	body, _ := json.Marshal(SignUpBody{"", "pass1234"})

	handler := NewSignUpHandler(context.Background(), *log.Default(), fakeRepo{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))

	handler.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != 400 {
		t.Fatal("StatusCode should be 400")
	}

	var ups toggles.Ups
	json.NewDecoder(result.Body).Decode(&ups)
	if ups.Msg != "Both email & password are required to be not empty" {
		t.Fatal("Wrong response body msg")
	}
}
