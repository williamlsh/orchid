package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func TestMiddleware(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	s := Service{
		logger: zap.NewExample().Sugar(),
	}
	r := s.createServeMux()
	r.(*mux.Router).HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("It just panics!")
	})
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
