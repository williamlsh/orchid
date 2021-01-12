package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	ConfigOptions
	logger *zap.SugaredLogger
	cache  cache.Cache
	db     database.Database
}

// ConfigOptions provides all config options frontend service needs.
type ConfigOptions struct {
	FrontendHostPort string
	AuthSecrets      auth.ConfigOptions
	Email            email.ConfigOptions
}

// NewServer creates a new frontend.Server
func NewServer(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	config ConfigOptions,
) *Server {
	return &Server{
		config,
		logger,
		cache,
		db,
	}
}

// Run starts the frontend server
func (s *Server) Run() error {
	mux := s.createServeMux()
	return http.ListenAndServe(s.FrontendHostPort, mux)
}

// createServeMux registers all routers.
func (s *Server) createServeMux() http.Handler {
	r := mux.NewRouter()
	r.Use(s.Middleware)

	sr := r.PathPrefix("/api").Subrouter()

	// Routers of authentication.
	auth.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, sr)

	return r
}

// Middleware implements mux.Middleware. It's a general recovery middleware to catch all panics in every route.
func (s *Server) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Recovered in top route:", r)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
