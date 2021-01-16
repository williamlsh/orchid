package auth

import (
	"context"
	"net/http"

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
	logger  *zap.SugaredLogger
	cache   cache.Cache
	secrets ConfigOptions

	AccessUUID    string
	ForgeedUserID uint64 // ForgeedUserID is a forged id responding to frontend
}

// New returns a new AuthenticationMiddleware
func New(logger *zap.SugaredLogger, cache cache.Cache, secrets ConfigOptions) AuthenticationMiddleware {
	return AuthenticationMiddleware{
		logger:  logger,
		cache:   cache,
		secrets: secrets,
	}
}

// Middleware Implements mux.Middleware, which will be called for each request that needs authentication.
func (amw *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
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
	realID, err := confuse.DecodeID(amw.ForgeedUserID)
	if err != nil {
		panic(err)
	}
	return realID
}

// GetForgedUserID returns a parsed and existing user's id in forged format.
func (amw *AuthenticationMiddleware) GetForgedUserID() uint64 {
	return amw.ForgeedUserID
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
