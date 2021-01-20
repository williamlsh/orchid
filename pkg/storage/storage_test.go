package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/ossm-org/orchid/pkg/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestS3Client(t *testing.T) {
	ctx := logging.WithLogger(context.Background(), zap.NewExample().Sugar())

	// Start a Minio service before run this test.
	s3Client := New(ctx, ConfigOptions{
		Endpoint: "127.0.0.1:9000",
		ID:       "AKIAIOSFODNN7EXAMPLE",
		Secret:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Secure:   false,
	})

	// Create bucket.
	bucketName := "my-bucketname"
	objectName := "my-objectname"
	if err := s3Client.PrepareBuckets(ctx, bucketName); err != nil {
		t.Fatal(err)
	}

	// Prepare object.
	object, err := os.Open("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()

	objectStat, err := object.Stat()
	if err != nil {
		t.Fatal(err)
	}
	size := objectStat.Size()

	// Store object in bucket.
	result, err := s3Client.PutObject(ctx, bucketName, objectName, object, size, minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Uploaded=%s result=%+v\n", objectName, result)

	t.Run("get object from presigned url", func(t *testing.T) {
		// Retrieve object in bucket.
		reqParams := make(url.Values)
		reqParams.Set("response-content-disposition", "attachment; filename=\"Makefile\"")

		presignedURL, err := s3Client.PresignedGetObject(ctx, bucketName, objectName, 3*time.Minute, reqParams)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Presigned URL:", presignedURL)

		resp, err := http.Get(presignedURL.String())
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
			t.Fatalf("Response not ok, %s", resp.Status)
		}
		fmt.Printf("Response headers: %+v\n", resp.Header)

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, int(size), len(b))
	})

	t.Run("put object from presigned url", func(t *testing.T) {
		presignedURL, err := s3Client.PresignedPutObject(ctx, bucketName, objectName, 3*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Presigned URL:", presignedURL)

		path := "../../Makefile"
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		defer writer.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPut, presignedURL.String(), body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
			t.Fatalf("Response not ok, status=%s", resp.Status)
		}

		fmt.Printf("Response headers: %+v\n", resp.Header)
	})
}
