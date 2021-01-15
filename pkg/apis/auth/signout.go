package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
)

// SignOuter implements a sign out handler.
type SignOuter struct {
	logger  *zap.SugaredLogger
	db      database.Database
	cache   cache.Cache
	secrets ConfigOptions
}

// NewSignOuter returns a new SignOuter.
func NewSignOuter(logger *zap.SugaredLogger, db database.Database, cache cache.Cache, secrets ConfigOptions) SignOuter {
	return SignOuter{
		logger,
		db,
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
		if errors.Is(err, errTokenExpired) {
			httpx.FinalizeResponse(w, httpx.ErrAuthTokenExpired, nil)
			return
		}

		s.logger.Errorf("could not delete creds form cache: %v", err)
		httpx.FinalizeResponse(w, httpx.ErrUnauthorized, nil)
		return
	}

	op := mux.Vars(r)["operation"]
	if op == "deregister" {
		realUserUD, err := confuse.DecodeID(ids.ID)
		if err != nil {
			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}

		// Deregister user from database, if error occurs, it must be already deregisterd.
		if err := s.deregisterUserFromDatabase(r.Context(), realUserUD); err != nil {
			s.logger.Errorf("could not deregister user: %v", err)

			httpx.FinalizeResponse(w, httpx.ErrAuthAlreadyDeregistered, nil)
			return
		}
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

func (s SignOuter) deregisterUserFromDatabase(ctx context.Context, userid uint64) error {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	sql := `
		UPDATE users
		SET deregistered = $1
		WHERE id = $2;
	`
	if _, err := conn.Exec(ctx, sql, true, userid); err != nil {
		return err
	}

	return nil
}
