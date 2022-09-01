package router

import (
	"net/http"
)

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

type Middleware interface {
	Apply(http.ResponseWriter, *http.Request) bool
}

func NewRouter() Router {
	return Router{http.NewServeMux(), []Middleware{}}
}

func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) Handle(path string, handler http.Handler) {
	middlewares := r.middlewares
	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		for _, m := range middlewares {
			if success := m.Apply(w, req); !success {
				return
			}
		}
		handler.ServeHTTP(w, req)
	})
}

func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(p, h)
}
