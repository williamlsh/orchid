package auth

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"go.uber.org/zap"
)

// Deregistor implements a deregister handler.
type Deregistor struct {
	logger  *zap.SugaredLogger
	db      database.Database
	cache   cache.Cache
	secrets ConfigOptions
}

// NewDeregistor returns a new Deregistor.
func NewDeregistor(logger *zap.SugaredLogger, cache cache.Cache, db database.Database, secrets ConfigOptions) Deregistor {
	return Deregistor{
		logger,
		db,
		cache,
		secrets,
	}
}

func (d Deregistor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := d.parseTokenFromRequest(r)
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
	if err := deleteCredsFromCache(d.cache, []string{userIDsInfo.UUID, refreshIDsInfo.UUID}); err != nil {
		d.logger.Errorf("could not delete creds form cache: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	realUserUD, err := confuse.DecodeID(userIDsInfo.ID)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrInternalServer, nil)
		return
	}

	// Deregister user from database, if error occurs, it must be already deregisterd.
	if d.deregisterUserFromDatabase(r.Context(), realUserUD); err != nil {
		d.logger.Errorf("could not deregister user: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrAuthAlreadyDeregistered, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

func (d Deregistor) parseTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	return request.ParseFromRequest(
		r,
		request.AuthorizationHeaderExtractor,
		func(t *jwt.Token) (interface{}, error) {
			return d.secrets.AccessSecret, nil
		},
		request.WithClaims(jwt.MapClaims{}),
		request.WithParser(&jwt.Parser{
			ValidMethods: []string{jwt.SigningMethodHS256.Alg()},
		}),
	)
}

func (d Deregistor) deregisterUserFromDatabase(ctx context.Context, userid uint64) error {
	conn, err := d.db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	sql := `
		UPDATE users
		SET deregistered = true,
		WHERE id = $1;
	`
	if _, err := conn.Exec(ctx, sql, userid); err != nil {
		return err
	}

	return nil
}
