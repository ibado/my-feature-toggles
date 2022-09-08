package router

import (
	"net/http"
)

type router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

type Middleware func(next http.Handler) http.Handler

func NewRouter() router {
	return router{http.NewServeMux(), []Middleware{}}
}

func (r *router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *router) Handle(path string, handler http.Handler) {
	r.registerHandler(path, handler)
}

func (r *router) HandleFunc(p string, handler func(http.ResponseWriter, *http.Request)) {
	r.registerHandler(p, http.HandlerFunc(handler))
}

func (r *router) registerHandler(path string, handler http.Handler) {
	middlewares := r.middlewares
	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		h := resolveHandler(middlewares, handler)
		h.ServeHTTP(w, req)
	})
}

func resolveHandler(middlewares []Middleware, h http.Handler) http.Handler {
	switch len(middlewares) {
	case 0:
		return h
	case 1:
		return middlewares[0](h)
	default:
		head := middlewares[0]
		tail := middlewares[1:]
		return head(resolveHandler(tail, h))
	}
}
