package toggles

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func RegisterHandlers(mux *http.ServeMux) {

}

type ToggleHandler struct {
	ctx    context.Context
	repo   ToggleRepo
	logger log.Logger
}

type Toggle struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type Ups struct {
	Msg string `json:"error"`
}

func NewHandler(ctx context.Context, repo ToggleRepo, logger log.Logger) ToggleHandler {
	return ToggleHandler{ctx, repo, logger}
}

func (h ToggleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		toggles, err := h.repo.GetAll(h.ctx)
		if err != nil {
			h.errorResponse(err, w)
			return
		}
		h.jsonResponse(toggles, http.StatusOK, w)
	case "PUT":
		defer req.Body.Close()
		var toggle Toggle
		err := json.NewDecoder(req.Body).Decode(&toggle)
		if err != nil || toggle.Id == "" || toggle.Value == "" {
			res := Ups{"Both 'id' and 'value' are required"}
			h.jsonResponse(res, http.StatusBadRequest, w)
			return
		}
		err = h.repo.Add(h.ctx, toggle.Id, toggle.Value)
		if err != nil {
			h.errorResponse(err, w)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case "DELETE":
		id := strings.Replace(req.URL.Path, "/toggles/", "", -1)

		if len(id) == 0 || id == req.URL.Path {
			res := Ups{"A valid id is required for removing a toggle"}
			h.jsonResponse(res, http.StatusBadRequest, w)
			return
		}

		exist, err := h.repo.Exist(h.ctx, id)
		if err != nil {
			h.errorResponse(err, w)
			return
		}
		if !exist {
			msg := fmt.Sprintf("the id '%s' doesn't match with an existing toggle", id)
			h.jsonResponse(Ups{msg}, http.StatusBadRequest, w)
			return
		}

		err = h.repo.Remove(h.ctx, id)
		if err != nil {
			h.errorResponse(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		h.jsonResponse(Ups{"Method not allowed"}, http.StatusMethodNotAllowed, w)
	}
}

func (c ToggleHandler) errorResponse(err error, w http.ResponseWriter) {
	c.logger.Println("Error: " + err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}

func (c ToggleHandler) jsonResponse(response any, statusCode int, w http.ResponseWriter) {
	json, err := json.Marshal(response)
	if err != nil {
		c.errorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(json)
}
