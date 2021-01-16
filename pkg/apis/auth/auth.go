package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
	"go.uber.org/zap"
)

// Group groups all authentication routers.
func Group(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	email email.ConfigOptions,
	secrets ConfigOptions,
	r *mux.Router,
) {
	r.Handle("/signup", newSignUpper(logger, cache, db, email)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.Handle("/signin", newSignInner(logger, cache, db, secrets)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.Handle("/{operation:signout|deregister}", newSignOuter(logger, db, cache, secrets)).
		Methods(http.MethodGet)

	r.Handle("/token/refresh", newRefresher(logger, cache, secrets)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	// The account handler need authentication middleware while others don't need,
	// so we use a new subrouter in order to not affect other handlers.
	sr := r.NewRoute().Subrouter()
	amw := New(logger, cache, secrets)

	sr.Handle("/token/refresh", newAccount(logger, amw, cache, db, secrets, email)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")
}
