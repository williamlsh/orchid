package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ossm-org/orchid/pkg/cache"
	"go.uber.org/zap"
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
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify the token.
	token, err := verifyToken(reqBody.RefreshToken, rf.secrets.RefreshSecret)
	// If there is an error, the token must have expired.
	if err != nil {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}
	// Is token valid?
	if !tokenValid(token) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	idsInfo, err := extractTokenMetaData(token, kindRefreshCreds)
	if errors.Is(err, ErrTokenExpired) {
		http.Error(w, ErrRefreshTokenExpired.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	// Delete the previous Refresh Token
	deleted, err := rf.cache.CommonRedis.Del(idsInfo.UUID).Result()
	if err != nil || deleted == 0 {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Create new pairs of refresh and access tokens
	credentials, err := createCreds(idsInfo.UserID, rf.secrets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// Save the tokens metadata to redis
	if err := cacheCredential(idsInfo.UserID, credentials, rf.cache); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err := encodeCreds(w, credentials.AccessToken, credentials.RefreshToken, "Ok"); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}
}
