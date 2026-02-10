package apiserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"userService/internal/app/model"
	"userService/internal/app/store"

	"github.com/alexedwards/scs/v2"
)

type server struct {
	router         *http.ServeMux
	logger         *slog.Logger
	store          store.Store
	sessionManager *scs.SessionManager
}

var (
	errIncorrectEmailOrPassword = errors.New("incorrect Email or Password")
)

func newServer(store store.Store, sessionStore *scs.SessionManager) *server {
	s := &server{
		router:         http.NewServeMux(),
		logger:         slog.Default(),
		store:          store,
		sessionManager: sessionStore,
	}

	s.configureRouter()
	s.configureSessionManager()

	return s
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/users", s.handleUsersCreate())
	s.router.HandleFunc("/session", s.handleSessionCreate())
}

func (s *server) configureSessionManager() {
	s.sessionManager.Lifetime = 24 * time.Hour
	s.sessionManager.Cookie.Name = "session_id"
	s.sessionManager.Cookie.HttpOnly = true
	s.sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	s.sessionManager.Cookie.Secure = false
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := s.sessionManager.LoadAndSave(s.router)
	handler.ServeHTTP(w, r)
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

func (s *server) handleSessionCreate() http.HandlerFunc {
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

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		if err := s.sessionManager.RenewToken(r.Context()); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}

		s.sessionManager.Put(r.Context(), "user_id", u.ID)

		s.respond(w, r, http.StatusOK, nil)
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
