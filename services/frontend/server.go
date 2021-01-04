package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/services/cache"
	"go.uber.org/zap"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	ConfigOptions
	logger *zap.SugaredLogger
	cache  cache.Cache
}

// ConfigOptions provides all config options frontend service needs.
type ConfigOptions struct {
	FrontendHostPort string
	AuthSecrets      auth.ConfigOptions
	Email            email.ConfigOptions
}

// NewServer creates a new frontend.Server
func NewServer(logger *zap.SugaredLogger, cache cache.Cache, config ConfigOptions) *Server {
	return &Server{
		config,
		logger,
		cache,
	}
}

// Run starts the frontend server
func (s *Server) Run() error {
	mux := s.createServeMux()
	return http.ListenAndServe(s.FrontendHostPort, mux)
}

// createServeMux registers all routers.
func (s *Server) createServeMux() http.Handler {
	mux := mux.NewRouter()
	mux.Handle("/signup", auth.NewSignUpper(s.logger, s.cache, s.Email))
	mux.Handle("/signin", auth.NewSignInner(s.logger, s.cache, s.AuthSecrets))
	mux.Handle("/signout", auth.NewSignOuter(s.logger, s.cache))
	mux.Handle("/token/refresh", auth.NewRefresher(s.logger, s.cache, s.AuthSecrets))
	return mux
}
