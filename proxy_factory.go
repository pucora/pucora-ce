package velonetics

import (
	"fmt"

	cel "github.com/velonetics/velonetics-cel/v2"
	jsonschema "github.com/velonetics/velonetics-jsonschema/v2"
	lua "github.com/velonetics/velonetics-lua/v2/proxy"
	metrics "github.com/velonetics/velonetics-metrics/v2/gin"
	opencensus "github.com/velonetics/velonetics-opencensus/v2"
	"github.com/velonetics/lura/v2/config"
	"github.com/velonetics/lura/v2/logging"
	"github.com/velonetics/lura/v2/proxy"
)

func internalNewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory,
	metricCollector *metrics.Metrics) proxy.Factory {

	proxyFactory := proxy.NewDefaultFactory(backendFactory, logger)
	proxyFactory = proxy.NewShadowFactory(proxyFactory)
	proxyFactory = jsonschema.ProxyFactory(logger, proxyFactory)
	proxyFactory = cel.ProxyFactory(logger, proxyFactory)
	proxyFactory = lua.ProxyFactory(logger, proxyFactory)
	proxyFactory = metricCollector.ProxyFactory("pipe", proxyFactory)
	proxyFactory = opencensus.ProxyFactory(proxyFactory)
	return proxyFactory
}

// NewProxyFactory returns a new ProxyFactory wrapping the injected BackendFactory with the default proxy stack and a metrics collector
func NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
	proxyFactory := internalNewProxyFactory(logger, backendFactory, metricCollector)

	return proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the proxy pipe", cfg.Endpoint))
		return proxyFactory.New(cfg)
	})
}

type proxyFactory struct{}

func (proxyFactory) NewProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, metricCollector *metrics.Metrics) proxy.Factory {
	return NewProxyFactory(logger, backendFactory, metricCollector)
}
