package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"
	"github.com/williamlsh/orchid/pkg/apis/internal/httpx"
	"github.com/williamlsh/orchid/pkg/cache"
	"github.com/williamlsh/orchid/pkg/database"
	"github.com/williamlsh/orchid/pkg/email"
)

const (
	// cacheVerificationCodeKeyPrefix is an auth cache key prefix to set verification code.
	cacheVerificationCodeKeyPrefix = "auth:verification_code"

	verificationCodeLength     = 12
	verificationCodeExpiration = 2 * time.Hour

	operationRegister = "register"
	operationLogIn    = "login"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// signUpper implements a sign up handler.
// signUpper authenticates users by email thus combines both register and login operations
// and distinguishes these operatons from checking existing user or new user.
// It sends an authentication email to user.
type signUpper struct {
	logger   *zap.SugaredLogger
	mailConf email.ConfigOptions
	cache    cache.Cache
	db       database.Database
}

// newSignUpper returns a new SignUpper.
func newSignUpper(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	mailConf email.ConfigOptions,
) signUpper {
	return signUpper{
		logger,
		mailConf,
		cache,
		db,
	}
}

func (s signUpper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrRequestDecodeJSON, nil)
		return
	}

	if !isEmailValid(reqBody.Email) {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidEmail, nil)
		return
	}

	// Email letters should be lower case.
	lowercaseEmail := strings.ToLower(reqBody.Email)

	isNewUser, err := s.isNewUser(r.Context(), lowercaseEmail)
	if err != nil {
		s.logger.Errorf("could not check new user in database: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// Evict old code before cache new if any.
	if err := evictUserVerificationCode(r.Context(), s.cache, lowercaseEmail); err != nil && !errors.Is(err, redis.Nil) {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	code := randString(verificationCodeLength, letterBytes)
	if err := cacheUserEmail(r.Context(), s.cache, isNewUser, code, lowercaseEmail, verificationCodeExpiration); err != nil {
		s.logger.Errorf("could not cache verification code: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}
	s.logger.Debugf("Send email with isNewUser=%t token=%s", isNewUser, code)

	// Mark operation after caching new code.
	if err := markUserOperation(r.Context(), s.cache, lowercaseEmail, code, verificationCodeExpiration); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	subject, content, err := composeEmail(isNewUser, code)
	if err != nil {
		s.logger.Errorf("could not compose email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	mail := email.New(s.logger, s.mailConf, lowercaseEmail, subject)
	if err := mail.Send(content); err != nil {
		s.logger.Errorf("could not send code in email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

// cacheUserEmail caches user auth operation info with expiration value of verificationCodeExpiration.
// One user with unique email only has one cache info in redis.
// The SignInner handler will use cache info to authenticate user.
func cacheUserEmail(ctx context.Context, cache cache.Cache, isNewUser bool, code, email string, expiration time.Duration) error {
	// Embed code in cache key, then in /signin, we can construct this key from user submit code again.
	// Use email as value, then in /signin we can handle user with this email.
	key := cacheVerificationCodeKeyPrefix + ":" + code
	// Prefix operation to value.
	var val string

	if isNewUser {
		val = operationRegister + ":" + email
	} else {
		val = operationLogIn + ":" + email
	}

	return cache.Client.Set(ctx, key, val, expiration).Err()
}

// markUserOperation is an helper for cacheUserEmail.
// This helper marks user auth operation email with expiration value of verificationCodeExpiration.
// When user frequently request SignUpper handler to receive emails, we always mark the latest operation,
// delete the old cache, making only the latest operation is valid. This reduces SignInner handler complexity.
// When SignInner handler receives code from request, it handles only the latest verification code.
func markUserOperation(ctx context.Context, cache cache.Cache, email, code string, expiration time.Duration) error {
	key := cacheVerificationCodeKeyPrefix + ":" + email

	return cache.Client.Set(ctx, key, code, expiration).Err()
}

// evictUserVerificationCode is a helper for cacheUserEmail.
// It's called before cacheUserEmail.
func evictUserVerificationCode(ctx context.Context, cache cache.Cache, email string) error {
	// First, get code from email.
	key1 := cacheVerificationCodeKeyPrefix + ":" + email
	code, err := cache.Client.Get(ctx, key1).Result()
	if err != nil {
		// May return redis.Nil error, caller should check and skip this error.
		return err
	}

	// Second, delete code from code key.
	key2 := cacheVerificationCodeKeyPrefix + ":" + code
	return cache.Client.Del(ctx, key2).Err()
}

// isNewUser checks whether a signing up user is a new user by search its email in database.
func (s signUpper) isNewUser(ctx context.Context, email string) (bool, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var exists bool

	// An existing user has a email and doesn't deregister.
	sql := `select exists(select 1 from users where email = $1 and deregistered = $2)`
	if err := conn.QueryRow(ctx, sql, email, false).Scan(&exists); err != nil {
		return false, err
	}

	// An existing user is not a new user!
	return !exists, nil
}

// isEmailValid checks if the email provided passes the required structure
// and length test. It also checks the domain has a valid MX record.
func isEmailValid(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}
	if !emailRegex.MatchString(e) {
		return false
	}
	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

func composeEmail(isNewUser bool, code string) (subject string, content string, err error) {
	url := "https://example.com/m/callback?token=%s&operation=%s&state=example"
	if isNewUser {
		subject = "Finish creating your account on Example"
		regURL := fmt.Sprintf(url, code, "register")
		content, err = renderEmail(registerTpl, data{
			URL: template.URL(regURL),
		})
	} else {
		subject = "Sign in to Example"
		logURL := fmt.Sprintf(url, code, "login")
		content, err = renderEmail(loginTpl, data{
			URL: template.URL(logURL),
		})
	}
	return
}
