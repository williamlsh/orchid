package auth

import (
	"context"
	"net/http"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"go.uber.org/zap"
)

// AuthenticationMiddleware is a general JWT token validation,
// it also checks users in cache system.
type AuthenticationMiddleware struct {
	once    sync.Once
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions

	AccessUUID    string
	ForgeedUserID uint64 // ForgeedUserID is a forged id responding to frontend
	UserID        uint64
}

// New returns a new AuthenticationMiddleware
func New(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		logger:  logger,
		cache:   cache,
		secrets: secrets,
	}
}

// MiddlewareOptionallyAuthenticate can be used in route which optionally needs authentication.
// If a user is authenticated, this middleware parses the request token and extract user metedata.
// You can use either this middleware or MiddlewareMustAuthenticate but not both.
func (amw *AuthenticationMiddleware) MiddlewareOptionallyAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			amw.logger.Debug("A user without token")
			next.ServeHTTP(w, r)
		} else {
			amw.MiddlewareMustAuthenticate(next).ServeHTTP(w, r)
		}
	})
}

// MiddlewareMustAuthenticate Implements mux.MiddlewareMustAuthenticate, which will be called for each request that needs authentication.
func (amw *AuthenticationMiddleware) MiddlewareMustAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := request.ParseFromRequest(
			r,
			request.AuthorizationHeaderExtractor,
			func(t *jwt.Token) (interface{}, error) {
				return []byte(amw.secrets.AccessSecret), nil
			},
			request.WithClaims(jwt.MapClaims{}),
			request.WithParser(&jwt.Parser{
				ValidMethods: []string{jwt.SigningMethodHS256.Alg()},
			}),
		)
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

		exists, err := amw.isUserIDExistInCache(r.Context(), ids.UUID)
		if err != nil {
			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}
		if !exists {
			amw.logger.Debugf("Invalid user, access_uuid=%s forged_userid=%d", ids.UUID, ids.UserID)

			httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
			return
		}
		amw.logger.Debugf("Valid user, access_uuid=%s forged_userid=%d", ids.UUID, ids.UserID)

		amw.AccessUUID = ids.UUID
		amw.ForgeedUserID = ids.UserID

		next.ServeHTTP(w, r)
	})
}

// GetUserID returns a parsed and existing user's real id in databae.
func (amw *AuthenticationMiddleware) GetUserID() uint64 {
	amw.once.Do(func() {
		realID, err := confuse.DecodeID(amw.ForgeedUserID)
		if err != nil {
			panic(err)
		}
		amw.UserID = realID
	})
	return amw.UserID
}

// isUserExistInCache checks whether user id exists in cache.
// Any user's id that is not in cache system is not authenticated and valid.
func (amw *AuthenticationMiddleware) isUserIDExistInCache(ctx context.Context, uuid string) (bool, error) {
	n, err := amw.cache.Client.Exists(ctx, uuid).Result()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	return true, nil
}
