package toggles

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"myfeaturetoggles.com/toggles/util"
)

type toggleHandler struct {
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

func NewHandler(ctx context.Context, repo ToggleRepo, logger log.Logger) http.Handler {
	return toggleHandler{ctx, repo, logger}
}

func (h toggleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		toggles, err := h.repo.GetAll(h.ctx)
		if err != nil {
			util.ErrorResponse(err, w)
			return
		}
		util.JsonResponse(toggles, http.StatusOK, w)
	case "PUT":
		defer req.Body.Close()
		var toggle Toggle
		err := json.NewDecoder(req.Body).Decode(&toggle)
		if err != nil || toggle.Id == "" || toggle.Value == "" {
			res := Ups{"Both 'id' and 'value' are required"}
			util.JsonResponse(res, http.StatusBadRequest, w)
			return
		}
		err = h.repo.Add(h.ctx, toggle.Id, toggle.Value)
		if err != nil {
			util.ErrorResponse(err, w)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case "DELETE":
		id := strings.Replace(req.URL.Path, "/toggles/", "", -1)

		if len(id) == 0 || id == req.URL.Path {
			res := Ups{"A valid id is required: /toggles/<id>"}
			util.JsonResponse(res, http.StatusBadRequest, w)
			return
		}

		exist, err := h.repo.Exist(h.ctx, id)
		if err != nil {
			util.ErrorResponse(err, w)
			return
		}
		if !exist {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err = h.repo.Remove(h.ctx, id)
		if err != nil {
			util.ErrorResponse(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		util.JsonResponse(Ups{"Method not allowed"}, http.StatusMethodNotAllowed, w)
	}
}
