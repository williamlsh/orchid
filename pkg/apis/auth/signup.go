package auth

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/email"
	"go.uber.org/zap"
)

const verificationCodeKeyPrefix = "verification_code"

// SignUpper implements a sign up handler.
type SignUpper struct {
	logger   *zap.SugaredLogger
	mailConf email.Config
	cache    cache.Cache
}

// NewSignUpper returns a new SignUpper.
func NewSignUpper(logger *zap.SugaredLogger, mailConf email.Config, cache cache.Cache) SignUpper {
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
		email string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	code, err := encodeToString(6)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := s.cacheVerificationCode(code, reqBody.email); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mail := email.New(s.mailConf, reqBody.email, "Sign in to xxx")
	if err := mail.Send(code); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	io.WriteString(w, "Ok")
}

func encodeToString(max int) (string, error) {
	b := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, b, max)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func (s SignUpper) cacheVerificationCode(code, email string) error {
	return s.cache.Set(verificationCodeKeyPrefix+":"+email, code, "EX", 10*time.Minute)
}
