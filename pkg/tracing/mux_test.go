package tracing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

func TestTraceSingleRoute(t *testing.T) {
	tracer := Init("test", metrics.NullFactory, zap.NewExample().Sugar())

	myHandler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idstr := vars["productId"]
		//do something
		data := "Hello, we get " + idstr
		fmt.Fprintf(w, data)
	}

	r := mux.NewRouter()

	pattern := "/v1/products/{productId}"

	middleware := Middleware(
		tracer,
		http.HandlerFunc(myHandler),
	)

	r.Handle(pattern, middleware)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/products/7", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fail()
	}
}

func TestTraceAllRoutes(t *testing.T) {
	tracer := Init("test", metrics.NullFactory, zap.NewExample().Sugar())

	okHandler := func(w http.ResponseWriter, r *http.Request) {
		// do something
		data := "Hello"
		fmt.Fprintf(w, data)
	}

	r := mux.NewRouter()
	// Create multiples routes
	r.HandleFunc("/v1/products", okHandler)
	r.HandleFunc("/v1/products/{productId}", okHandler)
	r.HandleFunc("/v2/products", okHandler)
	r.HandleFunc("/v2/products/{productId}", okHandler)
	r.HandleFunc("/v3/products", okHandler)
	r.HandleFunc("/v3/products/{productId}", okHandler)
	r.HandleFunc("/v4/products", okHandler)
	r.HandleFunc("/v4/products/{productId}", okHandler)

	// Add tracing to all routes
	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		route.Handler(Middleware(tracer, route.GetHandler()))
		return nil
	})

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/products/7", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fail()
	}
}
