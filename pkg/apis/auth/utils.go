package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/ossm-org/orchid/pkg/cache"
)

func FetchCredsFromCache(uuid string, cache cache.Cache) (uint64, error) {
	// TODO: handle potential nil reply which expired.
	return redis.Uint64(cache.Get(uuid))
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	// Format: 'Authorization': 'Bearer <YOUR_TOKEN_HERE>'
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strings.TrimSpace(strArr[1])
	}
	return ""
}

func VerifyToken(r *http.Request, secret string) (*jwt.Token, error) {
	signedTok := ExtractToken(r)
	token, err := jwt.Parse(signedTok, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request, secret string) error {
	token, err := VerifyToken(r, secret)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}

// AccessCreds is access credentials.
type AccessCreds struct {
	UUID   string
	UserID uint64
}

func ExtractTokenMetaData(r *http.Request) (*AccessCreds, error) {
	token, err := VerifyToken(r, "")
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessCreds{
			UUID:   accessUUID,
			UserID: userID,
		}, nil
	}
	return nil, err
}
