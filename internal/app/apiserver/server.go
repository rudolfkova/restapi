package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"userService/internal/app/model"
	"userService/internal/app/store"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
)

type server struct {
	router         *http.ServeMux
	logger         *slog.Logger
	store          store.Store
	sessionManager *scs.SessionManager
}

const (
	ctxKeyUser ctxKey = iota
	ctsKeyRequest
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect Email or Password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type ctxKey int8

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

	private := http.NewServeMux()
	private.HandleFunc("/whoami", s.handleWoami())
	s.router.Handle("/private/", http.StripPrefix("/private", s.authenticateUser(private)))
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctsKeyRequest, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := s.logger.With("remote_addr", r.RemoteAddr, "request_id", r.Context().Value(ctsKeyRequest))
		log.Info("request started", "method", r.Method, "uri", r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Info("request completed", "duration", time.Since(start), "with code", rw.code, " ", http.StatusText(rw.code))
	})
}

func (s *server) configureSessionManager() {
	s.sessionManager.Lifetime = 24 * time.Hour
	s.sessionManager.Cookie.Name = "session_id"
	s.sessionManager.Cookie.HttpOnly = true
	s.sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	s.sessionManager.Cookie.Secure = false
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := s.sessionManager.GetInt(r.Context(), "user_id")
		if id == 0 {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := s.setRequestID(s.logRequest(cors(s.sessionManager.LoadAndSave(s.router))))
	handler.ServeHTTP(w, r)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleWoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
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
