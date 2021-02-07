package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/williamlsh/orchid/pkg/apis/auth"
	"github.com/williamlsh/orchid/pkg/apis/upload/v1"
	"github.com/williamlsh/orchid/pkg/apis/users"
	"github.com/williamlsh/orchid/pkg/cache"
	"github.com/williamlsh/orchid/pkg/database"
	"github.com/williamlsh/orchid/pkg/email"
	"github.com/williamlsh/orchid/pkg/storage"
	"github.com/williamlsh/orchid/pkg/tracing"
	"go.uber.org/zap"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	ConfigOptions
	logger  *zap.SugaredLogger
	tracer  opentracing.Tracer
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
	tracer opentracing.Tracer,
	cache cache.Cache,
	db database.Database,
	storage storage.S3Client,
	config ConfigOptions,
) *Server {
	return &Server{
		config,
		logger,
		tracer,
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
	r := mux.NewRouter()
	r.Use(s.Middleware)

	sr := r.PathPrefix("/api").Subrouter()

	// Routers of authentication.
	// We use subrouter in every mux group, so that every group can use their own middleware and doesn't effect other groups.
	auth.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, sr)

	// Routers of users. They are under /api/user
	userRouter := sr.PathPrefix("/user").Subrouter()
	users.Group(s.logger, s.cache, s.db, s.Email, s.AuthSecrets, userRouter)

	// Routers of upload. They are under /api/upload
	uploadRouter := sr.PathPrefix("/upload").Subrouter()
	upload.Group(s.logger, s.cache, s.storage, s.AuthSecrets, uploadRouter)

	// Opentracing for mux.
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		h := route.GetHandler()
		if h != nil {
			route.Handler(tracing.Middleware(s.tracer, route.GetHandler(), nethttp.MWComponentName("frontend")))
		}
		return nil
	})

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
