package apiserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"userService/internal/app/model"
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
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
