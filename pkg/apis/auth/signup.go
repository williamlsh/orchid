package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
)

const (
	// cacheVerificationCodeKeyPrefix is an auth cache key prefix to set verification code.
	cacheVerificationCodeKeyPrefix = "auth:verification_code"

	verificationCodeLength     = 12
	verificationCodeExpiration = 2 * time.Hour

	// letterBytes is used to generate random string.
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	operationRegister = "register"
	operationLogIn    = "login"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// SignUpper implements a sign up handler.
// SignUpper authenticates users by email thus combines both register and login operations
// and distinguishes these operatons from checking existing user or new user.
// It sends an authentication email to user.
type SignUpper struct {
	logger   *zap.SugaredLogger
	mailConf email.ConfigOptions
	cache    cache.Cache
	db       database.Database
}

// NewSignUpper returns a new SignUpper.
func NewSignUpper(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	mailConf email.ConfigOptions,
) SignUpper {
	return SignUpper{
		logger,
		mailConf,
		cache,
		db,
	}
}

func (s SignUpper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	isNewUser, err := s.isNewUser(r.Context(), reqBody.Email)
	if err != nil {
		s.logger.Errorf("could not check new user in database: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// Evict old code before cache new if any.
	if err := s.evictUserVerificationCode(r.Context(), reqBody.Email); err != nil && !errors.Is(err, redis.Nil) {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	code := randString(verificationCodeLength)
	if err := s.cacheUserEmail(r.Context(), isNewUser, code, reqBody.Email, verificationCodeExpiration); err != nil {
		s.logger.Errorf("could not cache verification code: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// Mark operation after caching new code.
	if err := s.markUserOperation(r.Context(), reqBody.Email, code, verificationCodeExpiration); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	subject, content, err := composeEmail(isNewUser, code)
	if err != nil {
		s.logger.Errorf("could not compose email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	mail := email.New(s.logger, s.mailConf, reqBody.Email, subject)
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
func (s SignUpper) cacheUserEmail(ctx context.Context, isNewUser bool, code, email string, expiration time.Duration) error {
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
	return s.cache.Client.Set(ctx, key, val, expiration).Err()
}

// markUserOperation is an helper for cacheUserEmail.
// This helper marks user auth operation email with expiration value of verificationCodeExpiration.
// When user frequently request SignUpper handler to receive emails, we always mark the latest operation,
// delete the old cache, making only the latest operation is valid. This reduces SignInner handler complexity.
// When SignInner handler receives code from request, it handles only the latest verification code.
func (s SignUpper) markUserOperation(ctx context.Context, email, code string, expiration time.Duration) error {
	key := cacheVerificationCodeKeyPrefix + ":" + email
	return s.cache.Client.Set(ctx, key, code, expiration).Err()
}

// evictUserVerificationCode is a helper for cacheUserEmail.
// It's called before cacheUserEmail.
func (s SignUpper) evictUserVerificationCode(ctx context.Context, email string) error {
	// First, get code from email.
	key1 := cacheVerificationCodeKeyPrefix + ":" + email
	code, err := s.cache.Client.Get(ctx, key1).Result()
	if err != nil {
		// May return redis.Nil error, caller should check and skip this error.
		return err
	}

	// Second, delete code from code key.
	key2 := cacheVerificationCodeKeyPrefix + ":" + code
	return s.cache.Client.Del(ctx, key2).Err()
}

// isNewUser checks whether a signing up user is a new user by search its email in database.
func (s SignUpper) isNewUser(ctx context.Context, email string) (bool, error) {
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

// randString returns a n length random string.
func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
	url := "https://overseastu.com/m/callback?token=%s&operation=%s&state=overseastu"
	if isNewUser {
		subject = "Finish creating your account on Overseastu"
		regURL := fmt.Sprintf(url, code, "register")
		content, err = renderEmail(registerTpl, data{
			URL: template.URL(regURL),
		})
	} else {
		subject = "Sign in to Overseastu"
		logURL := fmt.Sprintf(url, code, "login")
		content, err = renderEmail(loginTpl, data{
			URL: template.URL(logURL),
		})
	}
	return
}
