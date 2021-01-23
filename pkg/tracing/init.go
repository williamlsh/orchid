package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

// TODO: implement jaeger span logger.

// Init initializes a jaeger opentracer.
func Init(serviceName string, metricsFactory metrics.Factory, logger *zap.SugaredLogger) opentracing.Tracer {
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatalf("cannot parse Jaeger env vars: %v", err)
	}
	cfg.ServiceName = serviceName
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	jaegerLogger := jaegerLoggerAdapter{logger}

	metricsFactory = metricsFactory.Namespace(metrics.NSOptions{
		Name: serviceName,
		Tags: nil,
	})
	tracer, _, err := cfg.NewTracer(
		config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		logger.Fatalf("cannot initialize Jaeger Tracer: %v", err)
	}

	return tracer
}

// Compile time interface check.
var _ jaeger.Logger = jaegerLoggerAdapter{}

type jaegerLoggerAdapter struct {
	*zap.SugaredLogger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.Error(msg)
}
