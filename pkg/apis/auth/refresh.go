package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
)

// Refresher implements a token refresh handler.
type Refresher struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions
}

// NewRefresher returns a new Refresher.
func NewRefresher(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) Refresher {
	return Refresher{
		logger,
		cache,
		secrets,
	}
}

func (rf Refresher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := rf.parseTokenFromRequest(r)
	// If there is an error, the token must have expired.
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}
	// Is token valid?
	if !tokenValid(token) {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidToken, nil)
		return
	}

	// Extract token metadata. The error returned is invalid token.
	userIDsInfo, refreshIDsInfo, err := extractTokenIDsMetadada(token)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidToken, nil)
		return
	}

	// Delete old creds from cache, if error occurs, creds may not exist in cache.
	if err := deleteCredsFromCache(rf.cache, []string{userIDsInfo.UUID, refreshIDsInfo.UUID}); err != nil {
		rf.logger.Errorf("could not delete creds form cache: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	// Create new pairs of refresh and access tokens.
	credentials, err := createCreds(userIDsInfo.ID, rf.secrets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// Save the tokens metadata to redis.
	if err := cacheCredential(rf.cache, userIDsInfo.ID, credentials); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, map[string]string{
		"access_token":  credentials.AccessToken,
		"refresh_token": credentials.RefreshToken,
	})
}

// noop implements jwt request.Extractor interface.
type noop struct{}

func (n noop) ExtractToken(r *http.Request) (string, error) {
	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return "", err
	}
	return reqBody.RefreshToken, nil
}

func (rf Refresher) parseTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	var bodyExtractor noop
	return request.ParseFromRequest(
		r,
		bodyExtractor,
		func(t *jwt.Token) (interface{}, error) {
			return rf.secrets.AccessSecret, nil
		},
		request.WithClaims(jwt.MapClaims{}),
		request.WithParser(&jwt.Parser{
			ValidMethods: []string{jwt.SigningMethodHS256.Alg()},
		}),
	)
}
