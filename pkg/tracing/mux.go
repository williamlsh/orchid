package tracing

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
)

// Middleware is an opentracing middleware for gorilla mux router.
func Middleware(tr opentracing.Tracer, h http.Handler, options ...nethttp.MWOption) http.Handler {
	opNameFunc := func(r *http.Request) string {
		if route := mux.CurrentRoute(r); route != nil {
			if tpl, err := route.GetPathTemplate(); err == nil {
				return r.Proto + " " + r.Method + " " + tpl
			}
		}
		return r.Proto + " " + r.Method
	}
	var opts = []nethttp.MWOption{
		nethttp.OperationNameFunc(opNameFunc),
		nethttp.MWSpanObserver(func(span opentracing.Span, r *http.Request) {
			span.SetTag("http.remote-address", r.RemoteAddr)
			span.SetTag("http.content-type", r.Header.Get("Content-Type"))
			span.SetTag("http.content-length", r.Header.Get("Content-Length"))
		}),
	}
	opts = append(opts, options...)
	return nethttp.Middleware(tr, h, opts...)
}
