package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/apis/upload/v1"
	"github.com/ossm-org/orchid/pkg/apis/users"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/pkg/storage"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	ConfigOptions
	logger  *zap.SugaredLogger
	cache   cache.Cache
	db      database.Database
	storage storage.S3Client
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
	storage storage.S3Client,
	config ConfigOptions,
) *Server {
	return &Server{
		config,
		logger,
		cache,
		db,
		storage,
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
	mux.Use(s.Middleware)

	r := mux.PathPrefix("/api").Subrouter()

	// Routers of authentication.
	// We use subrouter in every mux group, so that every group can use their own middleware and doesn't effect other groups.
	auth.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, r)

	// Routers of users. They are under /api/user
	userRouter := r.PathPrefix("/user").Subrouter()
	users.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, userRouter)

	// Routers of upload. They are under /api/upload
	uploadRouter := r.PathPrefix("/upload").Subrouter()
	upload.Group(s.logger, s.cache, s.storage, s.AuthSecrets, uploadRouter)

	return mux
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
