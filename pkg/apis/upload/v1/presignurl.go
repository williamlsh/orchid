package upload

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/williamlsh/orchid/pkg/apis/internal/httpx"
	"github.com/williamlsh/orchid/pkg/storage"
	"go.uber.org/zap"
)

const (
	// PresignedPutObjectURLExpiration is expiration time for resigned put object url.
	PresignedPutObjectURLExpiration = 5 * time.Minute

	bucketName = "all"
)

type uploader struct {
	logger  *zap.SugaredLogger
	storage storage.S3Client
}

func newUploader(logger *zap.SugaredLogger, storage storage.S3Client) uploader {
	return uploader{
		logger,
		storage,
	}
}

// getPresignedPutObjectURL returns a presigned object put url to client,
// so client can perform an HTTP put request to store object files.
func (u uploader) getPresignedPutObjectURL() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody struct {
			Checksum string
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			httpx.FinalizeResponse(w, httpx.ErrRequestDecodeJSON, nil)
			return
		}
		if strings.TrimSpace(reqBody.Checksum) == "" {
			httpx.FinalizeResponse(w, httpx.ErrUploadEmptyChecksum, nil)
			return
		}

		url, err := u.storage.PresignedPutObject(r.Context(), bucketName, reqBody.Checksum, PresignedPutObjectURLExpiration)
		if err != nil {
			u.logger.Errorf("failed to get presigned object put url: %v", err)

			httpx.FinalizeResponse(w, httpx.ErrServiceUnavailable, nil)
			return
		}

		httpx.FinalizeResponse(w, httpx.Success, map[string]string{
			"presigned_url": url.String(),
		})
	})
}
