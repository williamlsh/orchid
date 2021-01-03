package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/email"
	"go.uber.org/zap"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	hostPort string
	logger   *zap.SugaredLogger
	cache    cache.Cache
}

// NewServer creates a new frontend.Server
func NewServer(logger *zap.SugaredLogger, cache cache.Cache) *Server {
	return &Server{
		hostPort: "",
		logger:   logger,
		cache:    cache,
	}
}

// Run starts the frontend server
func (s *Server) Run() error {
	mux := s.createServeMux()
	return http.ListenAndServe(s.hostPort, mux)
}

// createServeMux registers all routers.
func (s *Server) createServeMux() http.Handler {
	mux := mux.NewRouter()
	mux.Handle("/signup", auth.NewSignUpper(s.logger, email.Config{}, s.cache))
	mux.Handle("/signin", auth.NewSignInner(s.logger, s.cache, "", ""))
	mux.Handle("/signout", auth.NewSignOuter(s.logger, s.cache))
	return mux
}
