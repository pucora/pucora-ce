package pucora

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-contrib/uuid"
	"golang.org/x/sync/errgroup"

	pucorabf "github.com/pucora/bloomfilter/v2/pucora"
	asyncamqp "github.com/pucora/pucora-amqp/v2/async"
	asynckafka "github.com/pucora/pucora-pubsub/v2/async"
	kafkapkg "github.com/pucora/pucora-pubsub/v2/kafka"
	audit "github.com/pucora/pucora-audit"
	cel "github.com/pucora/pucora-cel/v2"
	cmd "github.com/pucora/pucora-cobra/v2"
	cors "github.com/pucora/pucora-cors/v2/gin"
	gelf "github.com/pucora/pucora-gelf/v2"
	gologging "github.com/pucora/pucora-gologging/v2"
	influxdb "github.com/pucora/pucora-influx/v2"
	jose "github.com/pucora/pucora-jose/v2"
	logstash "github.com/pucora/pucora-logstash/v2"
	metrics "github.com/pucora/pucora-metrics/v2/gin"
	opencensus "github.com/pucora/pucora-opencensus/v2"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/datadog"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/influxdb"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/jaeger"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/ocagent"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/prometheus"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/stackdriver"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/xray"
	_ "github.com/pucora/pucora-opencensus/v2/exporter/zipkin"
	kotel "github.com/pucora/pucora-otel"
	otellura "github.com/pucora/pucora-otel/lura"
	otelgin "github.com/pucora/pucora-otel/router/gin"
	usage "github.com/pucora/pucora-usage/v2"
	soap "github.com/pucora/pucora-soap/v2"
	maingrpc "github.com/pucora/pucora-grpc/v2"
	grpcserver "github.com/pucora/pucora-grpc/v2/server"
	"github.com/pucora/lura/v2/async"
	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/core"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/proxy"
	router "github.com/pucora/lura/v2/router/gin"
	"github.com/pucora/lura/v2/sd/dnssrv"
	serverhttp "github.com/pucora/lura/v2/transport/http/server"
	server "github.com/pucora/lura/v2/transport/http/server/plugin"
)

// NewExecutor returns an executor for the cmd package. The executor initalizes the entire gateway by
// registering the components and composing a RouterFactory wrapping all the middlewares.
func NewExecutor(ctx context.Context) cmd.Executor {
	eb := new(ExecutorBuilder)
	return eb.NewCmdExecutor(ctx)
}

// PluginLoader defines the interface for the collaborator responsible of starting the plugin loaders
// Deprecated: Use PluginLoaderWithContext
type PluginLoader interface {
	Load(folder, pattern string, logger logging.Logger)
}

// PluginLoaderWithContext defines the interface for the collaborator responsible of starting the plugin loaders
type PluginLoaderWithContext interface {
	LoadWithContext(ctx context.Context, folder, pattern string, logger logging.Logger)
}

// SubscriberFactoriesRegister registers all the required subscriber factories from the available service
// discover components and adapters and returns a service register function.
// The service register function will register the service by the given name and port to all the available
// service discover clients
type SubscriberFactoriesRegister interface {
	Register(context.Context, config.ServiceConfig, logging.Logger) func(string, int)
}

// TokenRejecterFactory returns a jose.ChainedRejecterFactory containing all the required jose.RejecterFactory.
// It also should setup and manage any service related to the management of the revocation process, if required.
type TokenRejecterFactory interface {
	NewTokenRejecter(context.Context, config.ServiceConfig, logging.Logger, func(string, int)) (jose.ChainedRejecterFactory, error)
}

// MetricsAndTracesRegister registers the defined observability components and returns a metrics collector,
// if required.
type MetricsAndTracesRegister interface {
	Register(context.Context, config.ServiceConfig, logging.Logger) *metrics.Metrics
}

// EngineFactory returns a gin engine, ready to be passed to the Pucora RouterFactory
type EngineFactory interface {
	NewEngine(config.ServiceConfig, router.EngineOptions) *gin.Engine
}

// ProxyFactory returns a Pucora proxy factory, ready to be passed to the Pucora RouterFactory
type ProxyFactory interface {
	NewProxyFactory(logging.Logger, proxy.BackendFactory, *metrics.Metrics) proxy.Factory
}

// BackendFactory returns a Pucora backend factory, ready to be passed to the Pucora proxy factory
type BackendFactory interface {
	NewBackendFactory(context.Context, logging.Logger, *metrics.Metrics) proxy.BackendFactory
}

// HandlerFactory returns a Pucora router handler factory, ready to be passed to the Pucora RouterFactory
type HandlerFactory interface {
	NewHandlerFactory(logging.Logger, *metrics.Metrics, jose.RejecterFactory) router.HandlerFactory
}

// LoggerFactory returns a Pucora Logger factory, ready to be passed to the Pucora RouterFactory
type LoggerFactory interface {
	NewLogger(config.ServiceConfig) (logging.Logger, io.Writer, error)
}

// RunServer defines the interface of a function used by the Pucora router to start the service
type RunServer func(context.Context, config.ServiceConfig, http.Handler) error

// RunServerFactory returns a RunServer with several wraps around the injected one
type RunServerFactory interface {
	NewRunServer(logging.Logger, router.RunServerFunc) RunServer
}

// AgentStarter defines a type that starts a set of agents
type AgentStarter interface {
	Start(
		context.Context,
		[]*config.AsyncAgent,
		logging.Logger,
		chan<- string,
		proxy.Factory,
	) func() error
}

// ExecutorBuilder is a composable builder. Every injected property is used by the NewCmdExecutor method.
type ExecutorBuilder struct {
	// PluginLoader is deprecated: Use PluginLoaderWithContext
	PluginLoader                PluginLoader
	PluginLoaderWithContext     PluginLoaderWithContext
	LoggerFactory               LoggerFactory
	SubscriberFactoriesRegister SubscriberFactoriesRegister
	TokenRejecterFactory        TokenRejecterFactory
	MetricsAndTracesRegister    MetricsAndTracesRegister
	EngineFactory               EngineFactory

	ProxyFactory        ProxyFactory
	BackendFactory      BackendFactory
	HandlerFactory      HandlerFactory
	RunServerFactory    RunServerFactory
	AgentStarterFactory AgentStarter

	Middlewares []gin.HandlerFunc
}

// NewCmdExecutor returns an executor for the cmd package. The executor initializes the entire gateway by
// delegating most of the tasks to the injected collaborators. They register the components and
// compose a RouterFactory wrapping all the middlewares.
// Every nil collaborator is replaced by the default one offered by this package.
func (e *ExecutorBuilder) NewCmdExecutor(ctx context.Context) cmd.Executor {
	e.checkCollaborators()

	return func(cfg config.ServiceConfig) {
		cfg.Normalize()

		logger, gelfWriter, gelfErr := e.LoggerFactory.NewLogger(cfg)
		if gelfErr != nil {
			return
		}

		if err := kafkapkg.ValidateConfig(&cfg); err != nil {
			logger.Error("[SERVICE: Kafka]", err.Error())
			return
		}

		if err := soap.ValidateConfig(&cfg); err != nil {
			logger.Error("[SERVICE: SOAP]", err.Error())
			return
		}

		logger.Info(fmt.Sprintf("Starting Pucora v%s", core.PucoraVersion))
		startReporter(ctx, logger, cfg)

		if wd, err := os.Getwd(); err == nil {
			logger.Info("Working directory is", wd)
		}

		dnssrv.SetTTL(cfg.DNSCacheTTL)

		if cfg.Plugin != nil {
			e.PluginLoaderWithContext.LoadWithContext(ctx, cfg.Plugin.Folder, cfg.Plugin.Pattern, logger)
		}

		metricCollector := e.MetricsAndTracesRegister.Register(ctx, cfg, logger)
		if metricsAndTracesCloser, ok := e.MetricsAndTracesRegister.(io.Closer); ok {
			defer metricsAndTracesCloser.Close()
		}

		// Initializes the global cache for the JWK clients if enabled in the config
		if err := jose.SetGlobalCacher(logger, cfg.ExtraConfig); err != nil && err != jose.ErrNoValidatorCfg {
			logger.Error("[SERVICE: JOSE]", err.Error())
		}
		tokenRejecterFactory, err := e.TokenRejecterFactory.NewTokenRejecter(
			ctx,
			cfg,
			logger,
			e.SubscriberFactoriesRegister.Register(ctx, cfg, logger),
		)
		if err != nil && err != pucorabf.ErrNoConfig {
			logger.Warning("[SERVICE: Bloomfilter]", err.Error())
		}

		grpcRegistry, grpcSvcCfg, err := maingrpc.Bootstrap(cfg, logger)
		if err != nil {
			logger.Error("[SERVICE: gRPC]", err.Error())
			return
		}
		if grpcRegistry != nil {
			logger.Info("[SERVICE: gRPC]", "catalog loaded")
		}

		bpf := e.BackendFactory.NewBackendFactory(ctx, logger, metricCollector)
		pf := e.ProxyFactory.NewProxyFactory(logger, bpf, metricCollector)
		// we move the proxy factory out of the default proxy factory to make
		// sure that is always the outer middleware and that wraps any internal
		// proxy layer middleware:
		pf = otellura.ProxyFactory(pf)

		agentPing := make(chan string, len(cfg.AsyncAgents))

		handlerF := e.HandlerFactory.NewHandlerFactory(logger, metricCollector, tokenRejecterFactory)
		handlerF = otelgin.New(handlerF)

		runServerChain := serverhttp.RunServerWithLoggerFactory(logger)
		runServerChain = otellura.GlobalRunServer(logger, runServerChain)
		runServerChain = grpcserver.RunServer(logger, grpcRegistry, grpcSvcCfg, pf, tokenRejecterFactory, runServerChain)
		runServerChain = router.RunServerFunc(e.RunServerFactory.NewRunServer(logger, runServerChain))

		// setup the pucora router
		routerFactory := router.NewFactory(router.Config{
			Engine: e.EngineFactory.NewEngine(cfg, router.EngineOptions{
				Logger: logger,
				Writer: gelfWriter,
				Health: (<-chan string)(agentPing),
			}),
			ProxyFactory:   pf,
			Middlewares:    e.Middlewares,
			Logger:         logger,
			HandlerFactory: handlerF,
			RunServer:      runServerChain,
		})

		// start the engines
		logger.Info("Starting the Pucora instance")

		if len(cfg.AsyncAgents) == 0 {
			routerFactory.NewWithContext(ctx).Run(cfg)
			return
		}

		// start the async agents in the same error group as the router
		g, gctx := errgroup.WithContext(ctx)
		gctx, closeGroupCtx := context.WithCancel(gctx)

		if cfg.SequentialStart {
			waitAgents := e.AgentStarterFactory.Start(gctx, cfg.AsyncAgents, logger, (chan<- string)(agentPing), pf)
			g.Go(waitAgents)
		} else {
			g.Go(func() error {
				return e.AgentStarterFactory.Start(gctx, cfg.AsyncAgents, logger, (chan<- string)(agentPing), pf)()
			})
		}

		g.Go(func() error {
			logger.Info("[SERVICE: Gin] Building the router")
			routerFactory.NewWithContext(gctx).Run(cfg)
			closeGroupCtx()
			return nil
		})

		if err := g.Wait(); err != nil {
			logger.Error("[SERVICE: Async]", err.Error())
		}
	}
}

func (e *ExecutorBuilder) checkCollaborators() {
	if e.PluginLoader == nil {
		e.PluginLoader = new(pluginLoader)
	}
	if e.PluginLoaderWithContext == nil {
		e.PluginLoaderWithContext = new(pluginLoader)
	}
	if e.SubscriberFactoriesRegister == nil {
		e.SubscriberFactoriesRegister = new(registerSubscriberFactories)
	}
	if e.TokenRejecterFactory == nil {
		e.TokenRejecterFactory = new(BloomFilterJWT)
	}
	if e.MetricsAndTracesRegister == nil {
		e.MetricsAndTracesRegister = new(MetricsAndTraces)
	}
	if e.EngineFactory == nil {
		e.EngineFactory = new(engineFactory)
	}
	if e.ProxyFactory == nil {
		e.ProxyFactory = new(proxyFactory)
	}
	if e.BackendFactory == nil {
		e.BackendFactory = new(backendFactory)
	}
	if e.HandlerFactory == nil {
		e.HandlerFactory = new(handlerFactory)
	}
	if e.LoggerFactory == nil {
		e.LoggerFactory = new(LoggerBuilder)
	}
	if e.RunServerFactory == nil {
		e.RunServerFactory = new(DefaultRunServerFactory)
	}
	if e.AgentStarterFactory == nil {
		e.AgentStarterFactory = async.AgentStarter([]async.Factory{asyncamqp.StartAgent, asynckafka.StartAgent})
	}
}

// DefaultRunServerFactory creates the default RunServer by wrapping the injected RunServer
// with the plugin loader and the CORS module
type DefaultRunServerFactory struct{}

func (*DefaultRunServerFactory) NewRunServer(l logging.Logger, next router.RunServerFunc) RunServer {
	return RunServer(server.New(
		l,
		server.RunServer(cors.NewRunServerWithLogger(cors.RunServer(next), l)),
	))
}

// LoggerBuilder is the default BuilderFactory implementation.
type LoggerBuilder struct{}

// NewLogger sets up the logging components as defined at the configuration.
func (LoggerBuilder) NewLogger(cfg config.ServiceConfig) (logging.Logger, io.Writer, error) {
	var writers []io.Writer
	gelfWriter, gelfErr := gelf.NewWriter(cfg.ExtraConfig)
	if gelfErr == nil {
		writers = append(writers, gelfWriterWrapper{gelfWriter})
		gologging.SetFormatterSelector(func(w io.Writer) string {
			switch w.(type) {
			case gelfWriterWrapper:
				return "%{message}"
			default:
				return gologging.DefaultPattern
			}
		})
	} else {
		gelfWriter = nil
	}

	logger, gologgingErr := logstash.NewLogger(cfg.ExtraConfig)

	if gologgingErr != nil {
		logger, gologgingErr = gologging.NewLogger(cfg.ExtraConfig, writers...)

		if gologgingErr != nil {
			var err error
			logger, err = logging.NewLogger("DEBUG", os.Stdout, "PUCORA")
			if err != nil {
				return logger, gelfWriter, err
			}
			if gologgingErr != gologging.ErrWrongConfig {
				logger.Error("[SERVICE: Logging] Unable to create the logger:", gologgingErr.Error())
			}
		}
	}

	if gelfErr != nil && gelfErr != gelf.ErrWrongConfig {
		logger.Error("[SERVICE: Logging][GELF] Unable to create the writer:", gelfErr.Error())
	}

	if gologgingErr == nil {
		logger.Debug("[SERVICE: telemetry/logging] Improved logging started.")
	}

	return logger, gelfWriter, nil
}

// BloomFilterJWT is the default TokenRejecterFactory implementation.
type BloomFilterJWT struct{}

// NewTokenRejecter registers the bloomfilter component and links it to a token rejecter. Then it returns a chained
// rejecter factory with the created token rejecter and other based on the CEL component.
func (BloomFilterJWT) NewTokenRejecter(ctx context.Context, cfg config.ServiceConfig, l logging.Logger, reg func(n string, p int)) (jose.ChainedRejecterFactory, error) {
	rejecter, err := pucorabf.Register(ctx, "pucora-bf", cfg, l, reg)

	return jose.ChainedRejecterFactory([]jose.RejecterFactory{
		jose.RejecterFactoryFunc(func(_ logging.Logger, _ *config.EndpointConfig) jose.Rejecter {
			return jose.RejecterFunc(rejecter.RejectToken)
		}),
		jose.RejecterFactoryFunc(func(l logging.Logger, cfg *config.EndpointConfig) jose.Rejecter {
			if r := cel.NewRejecter(l, cfg); r != nil {
				return r
			}
			return jose.FixedRejecter(false)
		}),
	}), err
}

// MetricsAndTraces is the default implementation of the MetricsAndTracesRegister interface.
type MetricsAndTraces struct {
	shutdownFn func()
}

// Register registers the metrics, influx and opencensus packages as required by the given configuration.
func (m *MetricsAndTraces) Register(ctx context.Context, cfg config.ServiceConfig, l logging.Logger) *metrics.Metrics {
	metricCollector := metrics.New(ctx, cfg.ExtraConfig, l)

	if err := influxdb.New(ctx, cfg.ExtraConfig, metricCollector, l); err != nil {
		if err != influxdb.ErrNoConfig {
			l.Warning("[SERVICE: InfluxDB]", err.Error())
		}
	} else {
		l.Debug("[SERVICE: InfluxDB] Service correctly registered")
	}

	if err := opencensus.Register(ctx, cfg, opencensus.DefaultViews...); err != nil {
		if err != opencensus.ErrNoConfig {
			l.Warning("[SERVICE: OpenCensus]", err.Error())
		}
	} else {
		l.Debug("[SERVICE: OpenCensus] Service correctly registered")
	}

	if shutdownFn, err := kotel.Register(ctx, l, cfg); err == nil {
		m.shutdownFn = shutdownFn
	} else {
		l.Error(fmt.Sprintf("[SERVICE: OpenTelemetry] cannot register exporters: %s", err.Error()))
	}

	return metricCollector
}

func (m *MetricsAndTraces) Close() {
	if m.shutdownFn != nil {
		m.shutdownFn()
	}
}

const (
	usageDisableEnv = "USAGE_DISABLE"
	usageConfigNS   = "telemetry/usage"
	usageDelay      = 5 * time.Second
)

func usageReportingDisabled(cfg config.ServiceConfig) bool {
	switch os.Getenv(usageDisableEnv) {
	case "1", "true", "TRUE":
		return true
	}

	usageCfg, ok := cfg.ExtraConfig[usageConfigNS].(map[string]interface{})
	if !ok {
		return false
	}
	if enabled, ok := usageCfg["enabled"].(bool); ok && !enabled {
		return true
	}
	if disabled, ok := usageCfg["disabled"].(bool); ok && disabled {
		return true
	}
	return false
}

func startReporter(ctx context.Context, logger logging.Logger, cfg config.ServiceConfig) {
	logPrefix := "[SERVICE: Telemetry]"
	if usageReportingDisabled(cfg) {
		return
	}

	clusterID, err := cfg.Hash()
	if err != nil {
		logger.Debug(logPrefix, "Unable to create the Cluster ID hash:", err.Error())
		return
	}

	go func() {
		time.Sleep(usageDelay)

		serverID := uuid.NewV4().String()
		logger.Debug(logPrefix, "Registering usage stats for Cluster ID", clusterID)

		s := audit.Parse(&cfg)
		a, _ := audit.Marshal(&s)
		if err := usage.Report(
			ctx,
			usage.Options{
				ClusterID:    clusterID,
				ServerID:     serverID,
				Version:      core.PucoraVersion,
				UserAgent:    core.PucoraUserAgent,
				ExtraPayload: a,
				Client:       &http.Client{Transport: serverhttp.NewTransport(cfg, logger)},
			},
			nil,
		); err != nil {
			logger.Debug(logPrefix, "Unable to create the usage report client:", err.Error())
		}
	}()
}

type gelfWriterWrapper struct {
	io.Writer
}
