package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"go.uber.org/zap"
)

// SignInner implements a sign in handler.
type SignInner struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	db      database.Database
	secrets ConfigOptions
}

// NewSignInner returns a new SignInner.
func NewSignInner(logger *zap.SugaredLogger, cache cache.Cache, db database.Database, secrets ConfigOptions) SignInner {
	return SignInner{
		logger,
		cache,
		db,
		secrets,
	}
}

func (s SignInner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email, Username, Code string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	code, err := s.fetchVerificationCodeFromCache(verificationCodeKeyPrefix + ":" + reqBody.Email)
	if err == redis.Nil {
		if err := encodeCreds(w, "", "", "Verification code expired"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if reqBody.Code != code {
		if err := encodeCreds(w, "", "", "Incorrect verification code"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		return
	}
	if err = s.deleteVerificationCodeFromCache(verificationCodeKeyPrefix + ":" + reqBody.Email); err != nil {
		s.logger.Warn("An error occurred when deleting cached verification code: %v", err)
	}

	userid, err := s.createUserIfNotExist(r.Context(), reqBody.Email, reqBody.Username)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.logger.Debugf("Created a new userid: %d\n", userid)

	forgedUserID, err := confuse.EncodeID(userid)
	if err != nil {
		s.logger.Errorf("could not forge userid: %v", forgedUserID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	credentials, err := createCreds(forgedUserID, s.secrets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	s.logger.Debugf("Credentials, access token: %s refresh token: %s access uuid: %s refresh uuid: %s access expired at: %d refresh expired at: %d\n", credentials.AccessToken, credentials.RefreshToken, credentials.AccessUUID, credentials.RefreshUUID, credentials.AccessExpireAt, credentials.RefreshExpireAt)

	if err := cacheCredential(forgedUserID, credentials, s.cache); err != nil {
		s.logger.Errorf("could not cache credentials: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := encodeCreds(w, credentials.AccessToken, credentials.RefreshToken, "Ok"); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}
}

func (s SignInner) fetchVerificationCodeFromCache(key string) (string, error) {
	return s.cache.Client.Get(key).Result()
}

func (s SignInner) deleteVerificationCodeFromCache(key string) error {
	_, err := s.cache.Client.Del(key).Result()
	return err
}

func (s SignInner) createUserIfNotExist(ctx context.Context, email, username string) (uint64, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var id uint64

	sql := `
	with a as (
		select id, email, username
		from users
		where email = $1 and username = $2
	), b as (
		insert into users (email, username)
		select $3, $4
		where not exists (
			select 1 from a
		)
		returning id
	)
	select id from a
	union all
	select id from b
	`
	if err := conn.QueryRow(ctx, sql, email, username, email, username).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
