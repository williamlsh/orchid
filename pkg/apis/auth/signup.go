package auth

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
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
		email string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	code := randString(12)
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s SignUpper) cacheVerificationCode(code, email string) error {
	return s.cache.Set(verificationCodeKeyPrefix+":"+email, code, "EX", 10*time.Minute)
}
