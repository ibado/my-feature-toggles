package toggles

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const fakeJwt = "header.eyJVc2VySWQiOjEwLCJJYXQiOjE2NjI4NTQ2NzB9.sign"

type FakeRepo struct {
	Err         error
	Entries     map[string]string
	ToggleExist bool
}

func (r FakeRepo) GetAll(ctx context.Context, userId int64) (map[string]string, error) {
	return r.Entries, r.Err
}

func (r FakeRepo) Add(ctx context.Context, id string, value string, userId int64) error {
	return r.Err
}

func (r FakeRepo) Remove(ctx context.Context, id string, userId int64) error {
	return r.Err
}

func (r FakeRepo) Exist(ctx context.Context, id string, userId int64) (bool, error) {
	return r.ToggleExist, r.Err
}

func TestGetTogglesSuccess(t *testing.T) {
	recorder := httptest.NewRecorder()

	toggleList := map[string]string{"id1": "value1", "id2": "value2"}
	request := httptest.NewRequest("GET", "/toggles", nil)
	request.Header.Add("Authorization", fakeJwt)
	repo := FakeRepo{Entries: toggleList}
	handler := NewHandler(context.Background(), repo, *log.Default())

	handler.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()

	var resBody []Toggle
	json.NewDecoder(result.Body).Decode(&resBody)

	expectedBody := []Toggle{{"id1", "value1"}, {"id2", "value2"}}
	if !reflect.DeepEqual(resBody, expectedBody) {
		t.Error("Response body doen't match")
	}
}

func TestPutTogglesSuccess(t *testing.T) {
	body := Toggle{"id", "value"}
	json, _ := json.Marshal(body)
	request := httptest.NewRequest("PUT", "/toggles", bytes.NewBuffer(json))
	request.Header.Add("Authorization", fakeJwt)
	recorder := httptest.NewRecorder()

	repo := FakeRepo{Err: nil}

	handler := NewHandler(context.Background(), repo, *log.Default())

	handler.ServeHTTP(recorder, request)

	result := recorder.Result()
	if result.StatusCode != http.StatusCreated {
		t.Error("Status code should be 201")
	}
}

func TestPutTogglesFail(t *testing.T) {
	toggle := Toggle{"", "value"}
	jsonBody, _ := json.Marshal(toggle)
	request := httptest.NewRequest("PUT", "/toggles", bytes.NewBuffer(jsonBody))
	request.Header.Add("Authorization", fakeJwt)
	recorder := httptest.NewRecorder()
	defer request.Body.Close()

	repo := FakeRepo{Err: nil}

	h := NewHandler(context.Background(), repo, *log.Default())

	h.ServeHTTP(recorder, request)

	result := recorder.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusBadRequest {
		t.Error("Status code should be 400")
	}
	var resBody map[string]string
	json.NewDecoder(result.Body).Decode(&resBody)
	if resBody["error"] != "Both 'id' and 'value' are required" {
		t.Error("Body msg should doesn't match")
	}
}

func TestDeleteToggleSuccess(t *testing.T) {
	toggleId := "someId"
	request := httptest.NewRequest(
		"DELETE",
		"/toggles/"+toggleId,
		nil,
	)
	request.Header.Add("Authorization", fakeJwt)
	recorder := httptest.NewRecorder()

	repo := FakeRepo{Err: nil, ToggleExist: true}
	handler := NewHandler(context.Background(), repo, *log.Default())

	handler.ServeHTTP(recorder, request)
	result := recorder.Result()

	if result.StatusCode != 200 {
		t.Error("Status code should be 200")
	}
}

func TestDeleteToggleFail(t *testing.T) {
	toggleId := ""
	request := httptest.NewRequest(
		"DELETE",
		"/toggles/"+toggleId,
		nil,
	)
	request.Header.Add("Authorization", fakeJwt)
	recorder := httptest.NewRecorder()

	repo := FakeRepo{Err: nil, ToggleExist: true}
	handler := NewHandler(context.Background(), repo, *log.Default())

	handler.ServeHTTP(recorder, request)
	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != 400 {
		t.Error("Status code should be 400")
	}

	var resBody map[string]string
	json.NewDecoder(result.Body).Decode(&resBody)

	if resBody["error"] != "A valid id is required: /toggles/<id>" {
		t.Error("response body msg should match: " + resBody["error"])
	}
}

func TestDeleteToggleFailWith404(t *testing.T) {
	toggleId := "id123"
	request := httptest.NewRequest(
		"DELETE",
		"/toggles/"+toggleId,
		nil,
	)
	request.Header.Add("Authorization", fakeJwt)
	recorder := httptest.NewRecorder()

	repo := FakeRepo{Err: nil, ToggleExist: false}
	handler := NewHandler(context.Background(), repo, *log.Default())

	handler.ServeHTTP(recorder, request)
	result := recorder.Result()

	if result.StatusCode != 404 {
		t.Error("Status code should be 404")
	}
}
