package users

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/williamlsh/orchid/pkg/apis/auth"
	"github.com/williamlsh/orchid/pkg/cache"
	"github.com/williamlsh/orchid/pkg/database"
	"github.com/williamlsh/orchid/pkg/email"
	"go.uber.org/zap"
)

// Group groups all authentication routers.
func Group(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	db database.Database,
	email email.ConfigOptions,
	secrets auth.ConfigOptions,
	r *mux.Router,
) {
	amw := auth.New(logger, cache, secrets)
	r.Use(amw.MiddlewareMustAuthenticate)

	// The user profile handlers.
	p := newProfile(logger, amw, db)

	r.HandleFunc("/profile", p.updateProfile()).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	r.HandleFunc("/profile", p.getProfile()).
		Methods(http.MethodGet)
}
