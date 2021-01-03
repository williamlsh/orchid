package auth

import (
	"io"
	"net/http"

	"github.com/ossm-org/orchid/pkg/cache"
	"go.uber.org/zap"
)

// SignOuter implements a sign out handler.
type SignOuter struct {
	logger *zap.SugaredLogger
	cache  cache.Cache
}

// NewSignOuter returns a new SignOuter.
func NewSignOuter(logger *zap.SugaredLogger, cache cache.Cache) SignOuter {
	return SignOuter{
		logger,
		cache,
	}
}

func (s SignOuter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessCreds, err := ExtractTokenMetaData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	deleted, err := s.cache.Delete(accessCreds.UUID)
	if err != nil || deleted == 0 {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	io.WriteString(w, "Successfully logged out")
}
