package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"github.com/williamlsh/orchid/pkg/apis/internal/httpx"
	"github.com/williamlsh/orchid/pkg/cache"
)

// refresher implements a token refresh handler.
type refresher struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions
}

// newRefresher returns a new Refresher.
func newRefresher(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) refresher {
	return refresher{
		logger,
		cache,
		secrets,
	}
}

func (rf refresher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := rf.parseTokenFromRequest(r)
	// If there is an error, the token must have expired.
	if err != nil {
		rf.logger.Errorf("failed to parse token: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}
	// Is token valid?
	if !tokenValid(token) {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidToken, nil)
		return
	}

	// Extract token metadata. The error returned is invalid token.
	refreshIDs, err := extractTokenMetaData(token, kindRefreshCreds)
	if err != nil {
		rf.logger.Errorf("failed to extract token metadata: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidToken, nil)
		return
	}

	// Delete old creds from cache, if error occurs, creds may not exist in cache.
	accessUUID := strings.Split(refreshIDs.UUID, "++")[0]
	if err := deleteCredsFromCache(r.Context(), rf.cache, []string{accessUUID, refreshIDs.UUID}); err != nil {
		rf.logger.Errorf("could not delete creds form cache: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	// Create new pairs of refresh and access tokens.
	credentials, err := createCreds(refreshIDs.UserID, rf.secrets)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}
	// Save the tokens metadata to redis.
	if err := cacheCredential(r.Context(), rf.cache, refreshIDs.UserID, credentials); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
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

func (rf refresher) parseTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	var bodyExtractor noop
	return request.ParseFromRequest(
		r,
		bodyExtractor,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(rf.secrets.RefreshSecret), nil
		},
		request.WithClaims(jwt.MapClaims{}),
		request.WithParser(&jwt.Parser{
			ValidMethods: []string{jwt.SigningMethodHS256.Alg()},
		}),
	)
}
