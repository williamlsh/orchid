package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
)

const verificationCodeKeyPrefix = "verification_code"

// SignUpper implements a sign up handler.
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
		httpx.FinalizeResponse(w, httpx.ErrInvalidEmail, nil)
		return
	}

	code := randString(12)
	if err := s.cacheVerificationCode(code, reqBody.Email, 2*time.Hour); err != nil {
		s.logger.Errorf("could not cache verification code: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrInternalServer, nil)
		return
	}

	isNewUser, err := s.isNewUser(r.Context(), reqBody.Email)
	if err != nil {
		s.logger.Errorf("could not check new user in database: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrInternalServer, nil)
		return
	}

	subject, content := composeEmail(isNewUser, code)

	mail := email.New(s.logger, s.mailConf, reqBody.Email, subject)
	if err := mail.Send(content); err != nil {
		s.logger.Errorf("could not send code in email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrInternalServer, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randString returns a n length random string.
func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s SignUpper) cacheVerificationCode(code, email string, expiration time.Duration) error {
	return s.cache.Client.Set(verificationCodeKeyPrefix+":"+email, code, expiration).Err()
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

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

// isNewUser checks whether a signing up user is a new user by search its email in database.
func (s SignUpper) isNewUser(ctx context.Context, email string) (bool, error) {
	conn, err := s.db.Pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var exists bool

	sql := `select exists(select 1 from users where email = $1)`
	if err := conn.QueryRow(ctx, sql, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func composeEmail(isNewUser bool, code string) (subject string, content string) {
	// TODO: Use a html template here.
	tpl := "https://overseastu.com/m/callback?token=%s&operation=%s&state=overseastu"
	switch isNewUser {
	case false:
		subject = "Sign in to Overseastu"
		content = fmt.Sprintf(tpl, code, "login")
	case true:
		subject = "Finish creating your account on Overseastu"
		content = fmt.Sprintf(tpl, code, "register")
	}
	return
}
