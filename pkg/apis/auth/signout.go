package auth

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
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
	token, err := s.parseTokenFromRequest(r)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}
	// Is token valid?
	if !tokenValid(token) {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidToken, nil)
		return
	}

	ids, err := extractTokenMetaData(token, kindAccessCreds)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	refreshUUID := ids.UUID + "++" + strconv.Itoa(int(ids.ID))
	if err := deleteCredsFromCache(r.Context(), s.cache, []string{ids.UUID, refreshUUID}); err != nil {
		s.logger.Errorf("could not delete creds form cache: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

func (s SignOuter) parseTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	return request.ParseFromRequest(
		r,
		request.AuthorizationHeaderExtractor,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(s.secrets.AccessSecret), nil
		},
		request.WithClaims(jwt.MapClaims{}),
		request.WithParser(&jwt.Parser{
			ValidMethods: []string{jwt.SigningMethodHS256.Alg()},
		}),
	)
}
