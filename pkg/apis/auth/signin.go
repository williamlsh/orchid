package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"go.uber.org/zap"
)

// SignInner implements a sign in handler.
// SignInner authenticates users by email thus combines both signup and signin operations
// and distinguishes these operatons from checking existing user or new user.
// It checks token in authentication email previously sent.
type SignInner struct {
	logger  *zap.SugaredLogger
	cache   cache.Cache
	db      database.Database
	secrets ConfigOptions
}

// NewSignInner returns a new SignInner.
func NewSignInner(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	secrets ConfigOptions,
) SignInner {
	return SignInner{
		logger,
		cache,
		db,
		secrets,
	}
}

func (s SignInner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The register operation submits alias in request body while login operation doesn't.
	var reqBody struct {
		Alias, Code, Operation string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrRequestDecodeJSON, nil)
		return
	}

	if len(reqBody.Code) != verificationCodeLength {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidVerificationCode, nil)
		return
	}

	// We compare reqBody.Code with cached verification code from redis to check validity and to determine operation type.
	// If reqBody.Code doesn't exist in database, just return soon.
	key := cacheVerificationCodeKeyPrefix + ":" + reqBody.Code
	val, err := s.fetchUserEmailFromCache(r.Context(), key)
	if errors.Is(err, redis.Nil) {
		// If key doesn't exist, verification code must be expired.
		httpx.FinalizeResponse(w, httpx.ErrAuthVerificationCodeExpired, nil)
		return
	}
	if err != nil {
		s.logger.Errorf("could not fetch email from cache: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// Delete caced verification code immediately once user requested,
	// so that a callack url in authentication email can be used only once.
	// This reduces complexity of bussiness logic.
	if err = s.deleteUserEmailFromCache(r.Context(), key); err != nil {
		s.logger.Errorf("An error occurred when deleting cached verification code: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// The cached verification code was fetched, then check user's operation.
	operation, email := splitOpAndEmail(val)
	if reqBody.Operation != operation {
		s.logger.Debugf("Operation in request is different from that associated with cached verification code, cached: %s, user: %s", reqBody.Operation, operation)

		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidOperation, nil)
		return
	}

	var userid uint64
	if operation == operationLogIn {
		userid, err = s.gerUserIDByEmail(r.Context(), email)
		if err != nil {
			s.logger.Errorf("could not get userid by email: %v", err)

			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}
		s.logger.Debugf("Got userid: %d\n", userid)
	} else if operation == operationRegister {
		// Handle empty user alias.
		if reqBody.Alias == "" {
			httpx.FinalizeResponse(w, httpx.ErrAuthEmptyAlias, nil)
			return
		}

		username, err := s.generateUsername(r.Context(), email)
		if err != nil {
			s.logger.Errorf("could not generate new username: %v", err)

			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}

		userid, err = s.createUser(r.Context(), email, username, reqBody.Alias)
		if err != nil {
			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}
		s.logger.Debugf("Created a new userid: %d\n", userid)
	}

	// Forge real userid from frontend.
	forgedUserID, err := confuse.EncodeID(userid)
	if err != nil {
		s.logger.Errorf("could not forge userid: %v", forgedUserID)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	credentials, err := createCreds(forgedUserID, s.secrets)
	if err != nil {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}
	s.logger.Debugf("Credentials, access token: %s refresh token: %s access uuid: %s refresh uuid: %s access expired at: %d refresh expired at: %d\n", credentials.AccessToken, credentials.RefreshToken, credentials.AccessUUID, credentials.RefreshUUID, credentials.AccessExpireAt, credentials.RefreshExpireAt)

	if err := cacheCredential(r.Context(), s.cache, forgedUserID, credentials); err != nil {
		s.logger.Errorf("could not cache credentials: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, map[string]string{
		"access_token":  credentials.AccessToken,
		"refresh_token": credentials.RefreshToken,
	})
}

func (s SignInner) fetchUserEmailFromCache(ctx context.Context, key string) (string, error) {
	return s.cache.Client.Get(ctx, key).Result()
}

func (s SignInner) deleteUserEmailFromCache(ctx context.Context, key string) error {
	_, err := s.cache.Client.Del(ctx, key).Result()
	return err
}

func (s SignInner) gerUserIDByEmail(ctx context.Context, email string) (uint64, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var id uint64

	sql := `
		SELECT "id" FROM "users" WHERE ("email" = $1)
	`
	if err := conn.QueryRow(ctx, sql, email).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// createUser creates a new user.
// A user has unique email and unique username but may not alias.
func (s SignInner) createUser(ctx context.Context, email, username, alias string) (uint64, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var id uint64

	sql := `
		insert into users(email, username, alias)
	`
	if err := conn.QueryRow(ctx, sql, email, alias, email, alias).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// splitOpAndEmail splits operation and email from cached verification value.
func splitOpAndEmail(val string) (operation string, email string) {
	subs := strings.Split(val, ":")
	return subs[0], subs[1]
}

// generateUsername generates a globally unique username.
// It generates username from email, if this usename is not unique,
// then it generates a random one.
func (s SignInner) generateUsername(ctx context.Context, email string) (string, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	var exists bool
	username := strings.Split(email, "@")[0]

	sql := `select exists(select 1 from users where username = $1)`
	if err := conn.QueryRow(ctx, sql, username).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return username, nil
	}

	// TODO: potential conflict if random string is still not unique.
	return randString(7), nil
}
