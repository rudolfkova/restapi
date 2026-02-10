package apiserver

import (
	"log/slog"
	"net/http"
	"userService/internal/app/store"
)

type server struct {
	router *http.ServeMux
	logger *slog.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router: http.NewServeMux(),
		logger: slog.Default(),
		store:  store,
	}

	s.configureRouter()

	return s
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/users", s.handleUsersCreate())
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){}
}
