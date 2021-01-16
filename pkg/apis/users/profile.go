package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/database"
	"go.uber.org/zap"
)

type userProfile struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type profile struct {
	logger *zap.SugaredLogger
	amw    auth.AuthenticationMiddleware
	db     database.Database
}

func newProfile(
	logger *zap.SugaredLogger,
	amw auth.AuthenticationMiddleware,
	db database.Database,
) profile {
	return profile{
		logger,
		amw,
		db,
	}
}

// updateProfile updates user's profile.
func (p profile) updateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: more user info can be added here later.
		var reqBody struct {
			Username string
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			httpx.FinalizeResponse(w, httpx.ErrRequestDecodeJSON, nil)
			return
		}

		if err := p.updateUsername(r.Context(), p.amw.GetUserID(), reqBody.Username); err != nil {
			p.logger.Errorf("failed to update username: %v", err)

			httpx.FinalizeResponse(w, httpx.ErrUsernameAlreadyInUse, nil)
			return
		}

		httpx.FinalizeResponse(w, httpx.Success, nil)
	}
}

// getProfile returns user's profile.
func (p profile) getProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := p.amw.GetUserID()
		userProfile, err := p.getUsername(r.Context(), userID)
		if err != nil {
			p.logger.Errorf("failed to get username, userid=%d, err=%v", userID, err)

			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}

		httpx.FinalizeResponse(w, httpx.Success, userProfile)
	}
}

// updateUsername is an helper for updateProfile.
func (p profile) updateUsername(ctx context.Context, userid uint64, new string) error {
	sql := `
		UPDATE users
		SET username = $1
		WHERE userid = $2;
	`
	return p.db.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql, new, userid)
		return err
	})
}

// getUsername is an helper for getProfile.
func (p profile) getUsername(ctx context.Context, userid uint64) (*userProfile, error) {
	conn, err := p.db.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	var profile userProfile

	sql := `
		SELECT username, email
		FROM users
		WHERE userid = $1;
	`
	err = conn.QueryRow(ctx, sql, userid).Scan(&profile.Username, &profile.Email)
	return &profile, err
}
