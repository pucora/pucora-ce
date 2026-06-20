// Pucora-ce sets up a complete Pucora API Gateway ready to serve

package main

import (
	"context"
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	pucora "github.com/pucora/pucora-ce/v2"
	cmd "github.com/pucora/pucora-cobra/v2"
	flexibleconfig "github.com/pucora/pucora-flexibleconfig/v2"
	koanf "github.com/pucora/pucora-koanf"
	"github.com/pucora/lura/v2/config"
)

const (
	fcPartials  = "FC_PARTIALS"
	fcTemplates = "FC_TEMPLATES"
	fcSettings  = "FC_SETTINGS"
	fcPath      = "FC_OUT"
	fcEnable    = "FC_ENABLE"
)

//go:embed schema
var embedSchema embed.FS

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case sig := <-sigs:
			log.Println("Signal intercepted:", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	pucora.RegisterEncoders()

	registerAliases()

	var cfg config.Parser
	cfg = koanf.New()
	if os.Getenv(fcEnable) != "" {
		cfg = flexibleconfig.NewTemplateParser(flexibleconfig.Config{
			Parser:    cfg,
			Partials:  os.Getenv(fcPartials),
			Settings:  os.Getenv(fcSettings),
			Path:      os.Getenv(fcPath),
			Templates: os.Getenv(fcTemplates),
		})
	}

	var rawSchema string
	schema, err := embedSchema.ReadFile("schema/schema.json")
	if err == nil {
		rawSchema = string(schema)
	}

	commandsToLoad := []cmd.Command{
		cmd.RunCommand,
		cmd.NewCheckCmd(rawSchema),
		cmd.PluginCommand,
		cmd.VersionCommand,
		cmd.AuditCommand,
		pucora.NewTestPluginCmd(),
	}

	cmd.DefaultRoot = cmd.NewRoot(cmd.RootCommand, commandsToLoad...)
	cmd.DefaultRoot.Cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.Execute(cfg, pucora.NewExecutor(ctx))
}

var aliases = map[string]string{
	"github_com/pucora/pucora/transport/http/server/handler":        "plugin/http-server",
	"github.com/pucora/pucora/transport/http/client/executor":       "plugin/http-client",
	"github.com/pucora/pucora/proxy/plugin":                         "plugin/req-resp-modifier",
	"github.com/pucora/pucora/proxy":                                "proxy",
	"github_com/pucora/lura/v2/router/gin":                              "router",

	"github.com/pucora/pucora-httpcache":                "qos/http-cache",
	"github.com/pucora/pucora-circuitbreaker/gobreaker": "qos/circuit-breaker",

	"github.com/pucora/pucora-apikeys":                  "auth/api-keys",
	"github.com/pucora/pucora-basicauth":                "auth/basic",
	"github.com/pucora/pucora-aws-sigv4":                "auth/aws-sigv4",
	"github.com/pucora/pucora-gcp-auth":                 "auth/gcp",
	"github.com/pucora/pucora-ntlm":                     "auth/ntlm",
	"github.com/pucora/pucora-oauth2-clientcredentials": "auth/client-credentials",
	"github.com/pucora/pucora-jose/validator":           "auth/validator",
	"github.com/pucora/pucora-jose/signer":              "auth/signer",
	"github_com/devopsfaith/bloomfilter":                        "auth/revoker",

	"github_com/pucora/pucora-botdetector": "security/bot-detector",
	"github_com/pucora/pucora-httpsecure":  "security/http",
	"github_com/pucora/pucora-cors":        "security/cors",

	"github.com/pucora/pucora-cel":        "validation/cel",
	"github.com/pucora/pucora-jsonschema": "validation/json-schema",

	"github.com/pucora/pucora-amqp/agent": "async/amqp",

	"github.com/pucora/pucora-amqp/consume":                  "backend/amqp/consumer",
	"github.com/pucora/pucora-amqp/produce":                  "backend/amqp/producer",
	"github.com/pucora/pucora-lambda":                        "backend/lambda",
	"github.com/pucora/pucora-soap/v2":                       "backend/soap",
	"github.com/pucora/pucora-grpc/v2":                       "grpc",
	"github.com/pucora/pucora-grpc/v2/client":                "backend/grpc",
	"github.com/pucora/pucora-pubsub/publisher":              "backend/pubsub/publisher",
	"github.com/pucora/pucora-pubsub/subscriber":             "backend/pubsub/subscriber",
	"github.com/pucora/pucora-pubsub/kafka/publisher":        "backend/pubsub/publisher/kafka",
	"github.com/pucora/pucora-pubsub/kafka/subscriber":       "backend/pubsub/subscriber/kafka",
	"github.com/pucora/pucora-pubsub/async":                "async/kafka",
	"github.com/pucora/lura/v2/transport/http/client/graphql": "backend/graphql",
	"github.com/pucora/pucora/http":                                 "backend/http",

	"github_com/pucora/pucora-gelf":       "telemetry/gelf",
	"github_com/pucora/pucora-gologging":  "telemetry/logging",
	"github_com/pucora/pucora-logstash":   "telemetry/logstash",
	"github_com/pucora/pucora-metrics":    "telemetry/metrics",
	"github_com/pucora/pucora-influx":     "telemetry/influx",
	"github_com/pucora/pucora-opencensus": "telemetry/opencensus",

	"github.com/pucora/pucora-lua/router":        "modifier/lua-endpoint",
	"github.com/pucora/pucora-lua/proxy":         "modifier/lua-proxy",
	"github.com/pucora/pucora-lua/proxy/backend": "modifier/lua-backend",
	"github.com/pucora/pucora-martian":           "modifier/martian",

	"github.com/pucora/pucora-fastjson":    "encoding/fast-json",
	"github.com/pucora/pucora-no-redirect": "backend/no-redirect",

	"github.com/pucora/pucora-jmespath":          "modifier/jmespath",
	"github.com/pucora/pucora-security-policies": "security/policies",
	"github.com/pucora/pucora-wildcard":          "plugin/wildcard",
	"github.com/pucora/pucora-openapi":           "documentation/openapi",
	"github.com/pucora/pucora-postman":           "documentation/postman",
	"github.com/pucora/pucora-middleware-plugin": "plugin/middleware",
}

func registerAliases() {
	for key, alias := range aliases {
		config.ExtraConfigAlias[alias] = key
	}

	// multi-namespace modules (same Go package, multiple config namespaces)
	config.ExtraConfigAlias["modifier/response-body"] = "github.com/pucora/pucora-response-body"
	config.ExtraConfigAlias["modifier/response-body-generator"] = "github.com/pucora/pucora-response-body"
	config.ExtraConfigAlias["modifier/request-body-extractor"] = "github.com/pucora/pucora-request-body"
	config.ExtraConfigAlias["modifier/request-body-generator"] = "github.com/pucora/pucora-request-body"
	config.ExtraConfigAlias["validation/response-json-schema"] = "github.com/pucora/pucora-jsonschema/v2"
	config.ExtraConfigAlias["qos/ratelimit/router/redis"] = "github.com/pucora/pucora-ratelimit/v3"
	config.ExtraConfigAlias["qos/ratelimit/tiered"] = "github.com/pucora/pucora-ratelimit/v3"
}
