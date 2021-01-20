package storage

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ossm-org/orchid/pkg/logging"
	"go.uber.org/zap"
)

// ConfigOptions is configurate options for Minio.
type ConfigOptions struct {
	Endpoint string
	ID       string
	Secret   string
	Secure   bool
}

// S3Client is a AWS S3 compatible Minio client.
type S3Client struct {
	*minio.Client
	logger *zap.SugaredLogger
}

// New returns a new S3Client.
func New(ctx context.Context, config ConfigOptions) S3Client {
	logger := logging.FromContext(ctx)

	s3Client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.ID, config.Secret, ""),
		Secure: config.Secure,
	})
	if err != nil {
		panic(err)
	}

	s3Client.TraceErrorsOnlyOn(os.Stderr)

	return S3Client{s3Client, logger}
}

// PrepareBuckets prepares buckets before object operations.
func (c S3Client) PrepareBuckets(ctx context.Context, buckets ...string) error {
	for _, b := range buckets {
		ok, err := c.BucketExists(context.Background(), b)
		if err != nil {
			return err
		}
		if !ok {
			// TODO: consider to enable object lock.
			if err := c.MakeBucket(ctx, b, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}
