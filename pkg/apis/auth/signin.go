package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/ossm-org/orchid/pkg/cache"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// SignInner implements a sign in handler.
type SignInner struct {
	logger                      *zap.SugaredLogger
	cache                       cache.Cache
	accessSecret, refreshSecret string
}

// NewSignInner returns a new SignInner.
func NewSignInner(logger *zap.SugaredLogger, cache cache.Cache, accessSecret, refreshSecret string) SignInner {
	return SignInner{
		logger,
		cache,
		accessSecret,
		refreshSecret,
	}
}

func (s SignInner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var reqBody struct {
		email, code string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	code, err := s.fetchVerificationCodeFromCache(verificationCodeKeyPrefix + reqBody.email)
	if err == redis.ErrNil {
		if err := encodeCreds(w, "", "", "Verification code expired"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if reqBody.code != code {
		if err := encodeCreds(w, "", "", "Incorrect verification code"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}

	credentials, err := s.createCreds(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := s.cacheCredential(0, credentials); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := encodeCreds(w, credentials.AccessToken, credentials.RefreshToken, "Ok"); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}
}

// CredsInfo is an authenticated user credentials collection.
type CredsInfo struct {
	AccessToken     string
	RefreshToken    string
	AccessUUID      string
	RefreshUUID     string
	AccessExpireAt  int64
	RefreshExpireAt int64
}

// createCreds creates JWT token with userid and secrets.
func (s SignInner) createCreds(userid uint64) (CredsInfo, error) {
	accessUUID := uuid.NewV4().String()
	refreshUUID := accessUUID + "++" + strconv.Itoa(int(userid))
	accessExpiredAt := time.Now().Add(time.Minute * 15).Unix()
	refreshExpiredAt := time.Now().Add(time.Hour * 24 * 7).Unix()

	accessClaims := jwt.MapClaims{
		"authorized":  true,
		"access_uuid": accessUUID,
		"user_id":     userid,
		"exp":         accessExpiredAt,
	}
	accessToken, err := createToken(accessClaims, s.accessSecret)
	if err != nil {
		return CredsInfo{}, err
	}

	refreshClaims := jwt.MapClaims{
		"refresh_uuid": refreshUUID,
		"user_id":      userid,
		"exp":          refreshExpiredAt,
	}
	refreshToken, err := createToken(refreshClaims, s.refreshSecret)
	if err != nil {
		return CredsInfo{}, err
	}

	return CredsInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessUUID:   accessUUID,
		RefreshUUID:  refreshUUID,
	}, nil
}

func createToken(claims jwt.MapClaims, secret string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return accessToken.SignedString([]byte(secret))
}

func (s SignInner) cacheCredential(userid uint64, creds CredsInfo) error {
	accessExpiredAt := time.Unix(creds.AccessExpireAt, 0)
	refreshExpiredAt := time.Unix(creds.RefreshExpireAt, 0)
	uid := strconv.Itoa(int(userid))
	now := time.Now()

	if err := s.cache.Set(creds.AccessUUID, uid, "EX", accessExpiredAt.Sub(now)); err != nil {
		return err
	}
	if err := s.cache.Set(creds.RefreshUUID, uid, "EX", refreshExpiredAt.Sub(now)); err != nil {
		return err
	}

	return nil
}

func (s SignInner) fetchVerificationCodeFromCache(email string) (string, error) {
	return redis.String(s.cache.Get(email))
}

func encodeCreds(w http.ResponseWriter, accessToken, refreshToken, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"msg":           msg,
	})
}
