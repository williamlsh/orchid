package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/ossm-org/orchid/pkg/apis/internal/confuse"
	"github.com/ossm-org/orchid/pkg/apis/internal/httpx"
	"github.com/ossm-org/orchid/pkg/cache"
	"go.uber.org/zap"
)

func TestMiddleware(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	secrets := ConfigOptions{
		AccessSecret:  "abc",
		RefreshSecret: "xyz",
	}
	cache := cache.Cache{Client: client}
	amw := New(
		zap.NewExample().Sugar(),
		cache,
		secrets,
	)

	forgedUserID, err := confuse.EncodeID(1)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := createCreds(forgedUserID, secrets)
	if err != nil {
		t.Fatal(err)
	}

	if err := cacheCredential(context.Background(), cache, forgedUserID, creds); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.AccessToken))

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Use(amw.Middleware)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.FinalizeResponse(w, httpx.Success, nil)
	})
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response httpx.FinalResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Code != httpx.Success {
		t.Errorf("returned: %s, want: %s", response.Code.Error(), httpx.Success.Error())
	}

	fmt.Printf("User-id: %d\n", amw.GetUserID())
}
