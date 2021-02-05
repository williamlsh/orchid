package upload

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/storage"
	"go.uber.org/zap"
)

// Group groups all authentication routers.
func Group(
	logger *zap.SugaredLogger,
	cache cache.Cache,
	storage storage.S3Client,
	secrets auth.ConfigOptions,
	r *mux.Router,
) {
	amw := auth.New(logger, cache, secrets)
	r.Use(amw.MiddlewareMustAuthenticate)

	uploader := newUploader(logger, storage)

	r.Handle("/upload_url", uploader.getPresignedPutObjectURL()).
		Methods(http.MethodGet)
}
