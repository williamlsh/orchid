package auth

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/services/cache"
	"go.uber.org/zap"
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
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Email string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !isEmailValid(reqBody.Email) {
		http.Error(w, ErrEmailInvalid.Error(), http.StatusNotAcceptable)
		return
	}

	code := randString(12)
	if err := s.cacheVerificationCode(code, reqBody.Email, 2*time.Hour); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mail := email.New(s.mailConf, reqBody.Email, "Sign in to xxx")
	if err := mail.Send(code); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	io.WriteString(w, "Ok")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s SignUpper) cacheVerificationCode(code, email string, expire time.Duration) error {
	return s.cache.Set(verificationCodeKeyPrefix+":"+email, code, "EX", expire)
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
