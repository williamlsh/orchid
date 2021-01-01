package frontend

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server implements jaeger-demo-frontend service
type Server struct {
	hostPort string
	logger   *zap.SugaredLogger
}

// NewServer creates a new frontend.Server
func NewServer(logger *zap.SugaredLogger) *Server {
	return &Server{
		hostPort: "",
		logger:   logger,
	}
}

// Run starts the frontend server
func (s *Server) Run() error {
	mux := s.createServeMux()
	return http.ListenAndServe(s.hostPort, mux)
}

func (s *Server) createServeMux() http.Handler {
	mux := mux.NewRouter()
	mux.Handle("", http.HandlerFunc(nil))
	return mux
}
