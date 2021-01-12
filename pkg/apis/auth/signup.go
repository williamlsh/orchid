package auth

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/email"
)

const verificationCodeKeyPrefix = "verification_code"

// SignUpper implements a sign up handler.
type SignUpper struct {
	logger   *zap.SugaredLogger
	mailConf email.ConfigOptions
	cache    cache.Cache
}

// NewSignUpper returns a new SignUpper.
func NewSignUpper(logger *zap.SugaredLogger, cache cache.Cache, mailConf email.ConfigOptions) SignUpper {
	return SignUpper{
		logger,
		mailConf,
		cache,
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

	mail := email.New(s.logger, s.mailConf, reqBody.Email, "Sign in to xxx")
	if err := mail.Send(code); err != nil {
		s.logger.Errorf("could not send code in email: %v", err)
		httpx.FinalizeResponse(w, httpx.ErrInternalServer, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
