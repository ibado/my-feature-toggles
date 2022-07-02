package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestJsonResponse(t *testing.T) {

	recorder := httptest.NewRecorder()
	expectedStatuscode := 201
	expectedBody := map[string]int{"key": 666}
	JsonResponse(expectedBody, expectedStatuscode, recorder)
	result := recorder.Result()
	defer result.Body.Close()

	contentType := result.Header.Get("Content-Type")
	expectedCT := "application/json; charset=utf-8"
	statusCode := result.StatusCode

	var body map[string]int
	json.NewDecoder(result.Body).Decode(&body)

	if contentType != expectedCT {
		t.Errorf("Content-type should be: '%s' but is '%s'", expectedCT, contentType)
	}

	if statusCode != expectedStatuscode {
		t.Errorf("StatusCode should be: '%d' but is '%d'", expectedStatuscode, statusCode)
	}

	if !reflect.DeepEqual(body, expectedBody) {
		t.Errorf("Body should be: '%v' but is '%v'", expectedBody, body)
	}
}

func TestErrorResponse(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)

	recorder := httptest.NewRecorder()

	err := errors.New("ups!")
	ErrorResponse(err, recorder)

	result := recorder.Result()
	expectedSC := http.StatusInternalServerError

	if result.StatusCode != expectedSC {
		t.Errorf("Body should be: '%v' but is '%v'", expectedSC, result.Status)
	}

	expectedLog := "Error: " + err.Error()
	actualLog := string(logOutput.Bytes())
	if !strings.Contains(actualLog, expectedLog) {
		t.Errorf("Body should be: '%v' but is '%v'", expectedLog, actualLog)
	}
}
