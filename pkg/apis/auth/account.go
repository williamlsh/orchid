package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/williamlsh/orchid/pkg/apis/internal/httpx"
	"github.com/williamlsh/orchid/pkg/cache"
	"github.com/williamlsh/orchid/pkg/database"
	"github.com/williamlsh/orchid/pkg/email"
	"go.uber.org/zap"
)

type account struct {
	logger   *zap.SugaredLogger
	amw      *AuthenticationMiddleware
	cache    cache.Cache
	db       database.Database
	secrets  ConfigOptions
	mailConf email.ConfigOptions
}

func newAccount(
	logger *zap.SugaredLogger,
	amw *AuthenticationMiddleware,
	cache cache.Cache,
	db database.Database,
	secrets ConfigOptions,
	mailConf email.ConfigOptions,
) account {
	return account{
		logger,
		amw,
		cache,
		db,
		secrets,
		mailConf,
	}
}

func (a account) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email string
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrRequestDecodeJSON, nil)
		return
	}

	if !isEmailValid(reqBody.Email) {
		httpx.FinalizeResponse(w, httpx.ErrAuthInvalidEmail, nil)
		return
	}

	// Email letters should be lower case.
	lowercaseEmail := strings.ToLower(reqBody.Email)
	// If failed to update email, it must be an existing email in users database.
	if err := a.updateUserEmail(r.Context(), a.amw.GetUserID(), lowercaseEmail); err != nil {
		a.logger.Errorf("failed to update user email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrAuthEmailAlreadyInUse, nil)
		return
	}

	// Evict old code before cache new if any.
	if err := evictUserVerificationCode(r.Context(), a.cache, lowercaseEmail); err != nil && !errors.Is(err, redis.Nil) {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	code := randString(verificationCodeLength, letterBytes)
	if err := cacheUserEmail(r.Context(), a.cache, false, code, lowercaseEmail, verificationCodeExpiration); err != nil {
		a.logger.Errorf("could not cache verification code: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}
	a.logger.Debugf("Send email with isNewUser=%t token=%s", false, code)

	// Mark operation after caching new code.
	if err := markUserOperation(r.Context(), a.cache, lowercaseEmail, code, verificationCodeExpiration); err != nil {
		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	// TODO: add a new user operation: update email, and also add a new email verification mail template. Currently use signin email template.
	subject, content, err := composeEmail(false, code)
	if err != nil {
		a.logger.Errorf("could not compose email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	mail := email.New(a.logger, a.mailConf, lowercaseEmail, subject)
	if err := mail.Send(content); err != nil {
		a.logger.Errorf("could not send code in email: %v", err)

		httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
		return
	}

	httpx.FinalizeResponse(w, httpx.Success, nil)
}

func (a account) updateUserEmail(ctx context.Context, userid uint64, new string) error {
	// Just update user email. If new email conflicts with others', it returns error due to unique constraint on email.
	sql := `
		UPDATE users
		SET email = $1
		WHERE id = $2;
	`
	return a.db.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql, new, userid)
		return err
	})
}
