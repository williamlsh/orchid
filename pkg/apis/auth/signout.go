package auth

import (
	"io"
	"net/http"

	"github.com/ossm-org/orchid/services/cache"
	"go.uber.org/zap"
)

// SignOuter implements a sign out handler.
type SignOuter struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions
}

// NewSignOuter returns a new SignOuter.
func NewSignOuter(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) SignOuter {
	return SignOuter{
		logger,
		cache,
		secrets,
	}
}

func (s SignOuter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessCreds, err := s.extractTokenMetaData(r)
	if err != nil {
		s.logger.Errorf("could not extract token: %v", err)
		http.Error(w, ErrAccessTokenInvalid.Error(), http.StatusUnauthorized)
		return
	}

	deleted, err := s.cache.Delete(accessCreds.UUID)
	if err != nil || deleted == 0 {
		http.Error(w, ErrPreviouslySignnedOutUser.Error(), http.StatusUnauthorized)
		return
	}

	io.WriteString(w, "Successfully logged out")
}

func (s SignOuter) extractTokenMetaData(r *http.Request) (*IDSInfo, error) {
	token, err := VerifyToken(r, s.secrets.AccessSecret)
	if err != nil {
		return nil, err
	}

	return extractTokenMetaData(token, kindAccessCreds)
}
