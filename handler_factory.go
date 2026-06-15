package velonetics

import (
	"fmt"

	botdetector "github.com/velonetics/velonetics-botdetector/v2/gin"
	jose "github.com/velonetics/velonetics-jose/v2"
	ginjose "github.com/velonetics/velonetics-jose/v2/gin"
	lua "github.com/velonetics/velonetics-lua/v2/router/gin"
	metrics "github.com/velonetics/velonetics-metrics/v2/gin"
	opencensus "github.com/velonetics/velonetics-opencensus/v2/router/gin"
	ratelimit "github.com/velonetics/velonetics-ratelimit/v3/router/gin"
	"github.com/velonetics/lura/v2/config"
	"github.com/velonetics/lura/v2/logging"
	"github.com/velonetics/lura/v2/proxy"
	router "github.com/velonetics/lura/v2/router/gin"
	"github.com/velonetics/lura/v2/transport/http/server"

	"github.com/gin-gonic/gin"
)

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := router.CustomErrorEndpointHandler(logger, server.DefaultToHTTPError)
	handlerFactory = ratelimit.NewRateLimiterMw(logger, handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)

	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the http handler", cfg.Endpoint))
		return handlerFactory(cfg, p)
	}
}

type handlerFactory struct{}

func (handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactory(l, m, r)
}
