package pucora

import (
	"fmt"

	cel "github.com/pucora/pucora-cel/v2"
	jsonschema "github.com/pucora/pucora-jsonschema/v2"
	lua "github.com/pucora/pucora-lua/v2/proxy"
	metrics "github.com/pucora/pucora-metrics/v2/gin"
	opencensus "github.com/pucora/pucora-opencensus/v2"
	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/proxy"
	jmespath "github.com/pucora/pucora-jmespath"
	responsebody "github.com/pucora/pucora-response-body"
	requestbody "github.com/pucora/pucora-request-body"
)

func internalNewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory,
	metricCollector *metrics.Metrics) proxy.Factory {

	proxyFactory := proxy.NewDefaultFactory(backendFactory, logger)
	proxyFactory = proxy.NewShadowFactory(proxyFactory)
	proxyFactory = jsonschema.ProxyFactory(logger, proxyFactory)
	proxyFactory = jsonschema.ResponseProxyFactory(logger, proxyFactory)
	proxyFactory = cel.ProxyFactory(logger, proxyFactory)
	proxyFactory = lua.ProxyFactory(logger, proxyFactory)
	proxyFactory = requestbody.ProxyFactory(proxyFactory)
	proxyFactory = responsebody.ProxyFactory(proxyFactory)
	proxyFactory = jmespath.ProxyFactory(proxyFactory)
	proxyFactory = metricCollector.ProxyFactory("pipe", proxyFactory)
	proxyFactory = opencensus.ProxyFactory(proxyFactory)
	return proxyFactory
}

// NewProxyFactory returns a new ProxyFactory wrapping the injected BackendFactory with the default proxy stack and a metrics collector
func NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
	proxyFactory := internalNewProxyFactory(logger, backendFactory, metricCollector)

	return proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the proxy pipe", cfg.Endpoint))

		if hasConditionalBackends(cfg.Backend) {
			return newConditionalProxy(logger, cfg, backendFactory), nil
		}

		return proxyFactory.New(cfg)
	})
}

func hasConditionalBackends(backends []*config.Backend) bool {
	for _, backend := range backends {
		if isConditionalBackend(backend) {
			return true
		}
	}
	return false
}

func isConditionalBackend(backend *config.Backend) bool {
	if backend.ExtraConfig == nil {
		return false
	}
	v, ok := backend.ExtraConfig["backend/conditional"]
	if !ok || v == nil {
		return false
	}
	cfg, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	strategy, _ := cfg["strategy"].(string)
	return strategy == "header" || strategy == "policy" || strategy == "fallback"
}

func newConditionalProxy(logger logging.Logger, cfg *config.EndpointConfig, backendFactory proxy.BackendFactory) proxy.Proxy {
	return newConditionalRouter(logger, cfg.Backend, backendFactory)
}

type proxyFactory struct{}

func (proxyFactory) NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
	return NewProxyFactory(logger, backendFactory, metricCollector)
}
