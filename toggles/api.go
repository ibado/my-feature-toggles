package toggles

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"myfeaturetoggles.com/toggles/util"
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
			res := Ups{"A valid id is required for removing a toggle"}
			util.JsonResponse(res, http.StatusBadRequest, w)
			return
		}

		exist, err := h.repo.Exist(h.ctx, id)
		if err != nil {
			util.ErrorResponse(err, w)
			return
		}
		if !exist {
			msg := fmt.Sprintf("the id '%s' doesn't match with an existing toggle", id)
			util.JsonResponse(Ups{msg}, http.StatusBadRequest, w)
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
