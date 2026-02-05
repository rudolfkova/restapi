package apiserver

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"userService/internal/app/store"
)

type APIServer struct {
	config *Config
	logger *slog.Logger
	router *http.ServeMux
	store  *store.Store
}

var programLevel = new(slog.LevelVar)

func New(config *Config) *APIServer {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	return &APIServer{
		config: config,
		logger: slog.Default(),
		router: http.NewServeMux(),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	var lvl slog.Level
	err := lvl.UnmarshalText([]byte(s.config.LogLevel))
	if err != nil {
		return err
	}
	programLevel.Set(lvl)

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/hello", s.handleHello())
}

func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store)

	if err := st.Open(); err != nil {
		return err
	}

	s.store = st

	return nil
}

func (s *APIServer) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}
