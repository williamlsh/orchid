package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/ossm-org/orchid/services/cache"
	"go.uber.org/zap"
)

// SignInner implements a sign in handler.
type SignInner struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions
}

// NewSignInner returns a new SignInner.
func NewSignInner(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) SignInner {
	return SignInner{
		logger,
		cache,
		secrets,
	}
}

func (s SignInner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var reqBody struct {
		email, code string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	code, err := s.fetchVerificationCodeFromCache(verificationCodeKeyPrefix + reqBody.email)
	if err == redis.ErrNil {
		if err := encodeCreds(w, "", "", "Verification code expired"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if reqBody.code != code {
		if err := encodeCreds(w, "", "", "Incorrect verification code"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}

	credentials, err := createCreds(0, s.secrets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := cacheCredential(0, credentials, s.cache); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := encodeCreds(w, credentials.AccessToken, credentials.RefreshToken, "Ok"); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}
}

func (s SignInner) fetchVerificationCodeFromCache(email string) (string, error) {
	return redis.String(s.cache.Get(email))
}
