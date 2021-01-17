package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/apis/users"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
)

// Service implements jaeger-demo-frontend service
type Service struct {
	ConfigOptions
	logger *zap.SugaredLogger
	cache  cache.Cache
	db     database.Database
}

// ConfigOptions provides all config options frontend service needs.
type ConfigOptions struct {
	EnableTLS        bool
	MaxConnections   int    // Server max connections
	Hostname         string // Host name for TlS server
	FrontendHostPort string
	AuthSecrets      auth.ConfigOptions
	Email            email.ConfigOptions
}

// NewService creates a new frontend.Server
func NewService(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	config ConfigOptions,
) *Service {
	return &Service{
		config,
		logger,
		cache,
		db,
	}
}

// Run starts the frontend server
func (s *Service) Run() error {
	mux := s.createServeMux()

	server := NewServer(s.FrontendHostPort, mux)
	server.MaxConnections = s.MaxConnections
	if s.EnableTLS {
		s.logger.Debugf("Enabled TLS, hostname=%s", s.Hostname)
		server.GetCertificate(s.Hostname)
	}
	s.logger.Debugf("Server starts, addr=%s max-conn=%d", s.FrontendHostPort, s.MaxConnections)
	return server.Start()
}

// createServeMux registers all routers.
func (s *Service) createServeMux() http.Handler {
	mux := mux.NewRouter()
	mux.Use(s.Middleware)

	r := mux.PathPrefix("/api").Subrouter()

	// Routers of authentication.
	// We use subrouter in every mux group, so that every group can use their own middleware and doesn't effect other groups.
	auth.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, r)

	// Routers of users. They are under /api/user
	userRouter := r.PathPrefix("/user").Subrouter()
	users.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, userRouter)

	return mux
}

// Middleware implements mux.Middleware. It's a general recovery middleware to catch all panics in every route.
func (s *Service) Middleware(next http.Handler) http.Handler {
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
