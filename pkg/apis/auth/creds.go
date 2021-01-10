package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"

	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/confuse"
)

// CredsPairInfo is an authenticated user credentials collection.
type CredsPairInfo struct {
	AccessToken     string
	RefreshToken    string
	AccessUUID      string
	RefreshUUID     string
	AccessExpireAt  int64
	RefreshExpireAt int64
}

// createCreds creates JWT token with userid and secrets.
func createCreds(userid uint64, secrets ConfigOptions) (*CredsPairInfo, error) {
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
	accessToken, err := createToken(accessClaims, secrets.AccessSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"refresh_uuid": refreshUUID,
		"user_id":      userid,
		"exp":          refreshExpiredAt,
	}
	refreshToken, err := createToken(refreshClaims, secrets.RefreshSecret)
	if err != nil {
		return nil, err
	}

	return &CredsPairInfo{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessUUID:      accessUUID,
		RefreshUUID:     refreshUUID,
		AccessExpireAt:  accessExpiredAt,
		RefreshExpireAt: refreshExpiredAt,
	}, nil
}

func createToken(claims jwt.MapClaims, secret string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return accessToken.SignedString([]byte(secret))
}

func cacheCredential(userid uint64, creds *CredsPairInfo, cache cache.Cache) error {
	accessExpiredAt := time.Unix(creds.AccessExpireAt, 0)
	refreshExpiredAt := time.Unix(creds.RefreshExpireAt, 0)
	uid := strconv.Itoa(int(userid))
	now := time.Now()

	if err := cache.Client.Set(creds.AccessUUID, uid, accessExpiredAt.Sub(now)).Err(); err != nil {
		return err
	}
	if err := cache.Client.Set(creds.RefreshUUID, uid, refreshExpiredAt.Sub(now)).Err(); err != nil {
		return err
	}

	return nil
}

func encodeCreds(w http.ResponseWriter, accessToken, refreshToken, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"msg":           msg,
	})
}

func FetchCredsFromCache(uuid string, cache cache.Cache) (uint64, error) {
	// TODO: handle potential nil reply which expired.
	return cache.Client.Get(uuid).Uint64()
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
	return verifyToken(signedTok, secret)
}

func verifyToken(signedTok, secret string) (*jwt.Token, error) {
	return jwt.Parse(signedTok, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func TokenValid(r *http.Request, secret string) error {
	token, err := VerifyToken(r, secret)
	if err != nil {
		return err
	}

	if !tokenValid(token) {
		return errors.New("Invalid token")
	}
	return nil
}

func tokenValid(token *jwt.Token) bool {
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return false
	}
	return true
}

type credsKind int

const (
	kindAccessCreds credsKind = iota
	kindRefreshCreds
)

func extractTokenMetaData(token *jwt.Token, kind credsKind) (*IDSInfo, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		switch kind {
		case kindAccessCreds:
			return readIDSInfoFromClaims(claims, "access_uuid")
		case kindRefreshCreds:
			return readIDSInfoFromClaims(claims, "refresh_uuid")
		}
	}
	return nil, ErrTokenExpired
}

func readIDSInfoFromClaims(claims jwt.MapClaims, uuidKind string) (*IDSInfo, error) {
	uuid, ok := claims[uuidKind].(string)
	if !ok {
		return nil, fmt.Errorf("No %s in claims", uuidKind)
	}
	userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
	if err != nil {
		return nil, err
	}

	forgedUserID, err := confuse.EncodeID(userID)
	if err != nil {
		return nil, fmt.Errorf("could not forge userid: %v", err)
	}

	return &IDSInfo{
		UUID:   uuid,
		UserID: forgedUserID,
	}, nil
}

// IDSInfo is either access info or refresh info.
type IDSInfo struct {
	UUID   string
	UserID uint64
}
