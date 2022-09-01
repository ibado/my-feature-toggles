package auth

import (
	"net/http"

	"myfeaturetoggles.com/toggles/router"
)

func AuthMiddleware() router.Middleware {
	return authMiddleware{}
}

type authMiddleware struct{}

func (a authMiddleware) Apply(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
