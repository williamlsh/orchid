package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
	"go.uber.org/zap"
)

// Auth groups all authentication routers.
func Auth(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	email email.ConfigOptions,
	secrets ConfigOptions,
	r *mux.Router,
) {
	r.Handle("/signup", NewSignUpper(logger, cache, email)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.Handle("/signin", NewSignInner(logger, cache, db, secrets)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.Handle("/signout", NewSignOuter(logger, cache, secrets)).
		Methods(http.MethodGet)

	r.Handle("/token/refresh", NewRefresher(logger, cache, secrets)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")
}
