package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFinalizeResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FinalizeResponse(w, Success, nil)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var got FinalResponse
	if err = json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	expected := FinalResponse{Code: Success, Msg: Msgs[Success]}
	if got != expected {
		t.Errorf("handler returned unexpected body: got %+v want %+v",
			got, expected)
	}
}
