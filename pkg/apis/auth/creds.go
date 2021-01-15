package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"

	"github.com/ossm-org/orchid/pkg/cache"
)

type credsKind int

const (
	kindAccessCreds credsKind = iota
	kindRefreshCreds
)

// IDs is either access ids or refresh ids.
type IDs struct {
	UUID string
	ID   uint64
}

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

func cacheCredential(ctx context.Context, cache cache.Cache, userid uint64, creds *CredsPairInfo) error {
	accessExpiredAt := time.Unix(creds.AccessExpireAt, 0)
	refreshExpiredAt := time.Unix(creds.RefreshExpireAt, 0)
	uid := strconv.Itoa(int(userid))
	now := time.Now()

	if err := cache.Client.Set(ctx, creds.AccessUUID, uid, accessExpiredAt.Sub(now)).Err(); err != nil {
		return err
	}
	if err := cache.Client.Set(ctx, creds.RefreshUUID, uid, refreshExpiredAt.Sub(now)).Err(); err != nil {
		return err
	}

	return nil
}

func tokenValid(token *jwt.Token) bool {
	if _, ok := token.Claims.(jwt.Claims); ok && token.Valid {
		return true
	}
	return false
}

func extractTokenMetaData(token *jwt.Token, kind credsKind) (*IDs, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		switch kind {
		case kindAccessCreds:
			return readIDSInfoFromClaims(claims, "access_uuid")
		case kindRefreshCreds:
			return readIDSInfoFromClaims(claims, "refresh_uuid")
		}
	}
	return nil, errors.New("invalid token")
}

func readIDSInfoFromClaims(claims jwt.MapClaims, uuidKind string) (*IDs, error) {
	uuid, ok := claims[uuidKind].(string)
	if !ok {
		return nil, fmt.Errorf("No %s in claims", uuidKind)
	}
	userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
	if err != nil {
		return nil, err
	}

	return &IDs{
		UUID: uuid,
		ID:   userID,
	}, nil
}

func extractTokenIDsMetadada(token *jwt.Token) (userIDs *IDs, refreshIDs *IDs, err error) {
	userIDs, err = extractTokenMetaData(token, kindAccessCreds)
	if err != nil {
		return
	}
	refreshIDs, err = extractTokenMetaData(token, kindRefreshCreds)
	return
}

func deleteCredsFromCache(ctx context.Context, cache cache.Cache, uuids []string) error {
	for _, id := range uuids {
		deleted, err := cache.Client.Del(ctx, id).Result()
		if err != nil {
			return err
		}
		if deleted == 0 {
			return errors.New("Credential already expired")
		}
	}
	return nil
}
