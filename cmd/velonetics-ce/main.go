// Velonetics-ce sets up a complete Velonetics API Gateway ready to serve

package main

import (
	"context"
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	velonetics "github.com/velonetics/velonetics-ce/v2"
	cmd "github.com/velonetics/velonetics-cobra/v2"
	flexibleconfig "github.com/velonetics/velonetics-flexibleconfig/v2"
	koanf "github.com/velonetics/velonetics-koanf"
	"github.com/velonetics/lura/v2/config"
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

	velonetics.RegisterEncoders()

	for key, alias := range aliases {
		config.ExtraConfigAlias[alias] = key
	}

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
		velonetics.NewTestPluginCmd(),
	}

	cmd.DefaultRoot = cmd.NewRoot(cmd.RootCommand, commandsToLoad...)
	cmd.DefaultRoot.Cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.Execute(cfg, velonetics.NewExecutor(ctx))
}

var aliases = map[string]string{
	"github_com/velonetics/velonetics-ce/v2/transport/http/server/handler":  "plugin/http-server",
	"github.com/velonetics/velonetics-ce/v2/transport/http/client/executor": "plugin/http-client",
	"github.com/velonetics/velonetics-ce/v2/proxy/plugin":                   "plugin/req-resp-modifier",
	"github.com/velonetics/velonetics-ce/v2/proxy":                          "proxy",
	"github_com/velonetics/lura/v2/router/gin":                              "router",

	"github.com/velonetics/velonetics-httpcache":                "qos/http-cache",
	"github.com/velonetics/velonetics-circuitbreaker/gobreaker": "qos/circuit-breaker",

	"github.com/velonetics/velonetics-oauth2-clientcredentials": "auth/client-credentials",
	"github.com/velonetics/velonetics-jose/validator":           "auth/validator",
	"github.com/velonetics/velonetics-jose/signer":              "auth/signer",
	"github_com/devopsfaith/bloomfilter":                        "auth/revoker",

	"github_com/velonetics/velonetics-botdetector": "security/bot-detector",
	"github_com/velonetics/velonetics-httpsecure":  "security/http",
	"github_com/velonetics/velonetics-cors":        "security/cors",

	"github.com/velonetics/velonetics-cel":        "validation/cel",
	"github.com/velonetics/velonetics-jsonschema": "validation/json-schema",

	"github.com/velonetics/velonetics-amqp/agent": "async/amqp",

	"github.com/velonetics/velonetics-amqp/consume":                  "backend/amqp/consumer",
	"github.com/velonetics/velonetics-amqp/produce":                  "backend/amqp/producer",
	"github.com/velonetics/velonetics-lambda":                        "backend/lambda",
	"github.com/velonetics/velonetics-soap/v2":                       "backend/soap",
	"github.com/velonetics/velonetics-pubsub/publisher":              "backend/pubsub/publisher",
	"github.com/velonetics/velonetics-pubsub/subscriber":             "backend/pubsub/subscriber",
	"github.com/velonetics/velonetics-ce/v2/transport/http/client/graphql": "backend/graphql",
	"github.com/velonetics/velonetics-ce/v2/http":                          "backend/http",

	"github_com/velonetics/velonetics-gelf":       "telemetry/gelf",
	"github_com/velonetics/velonetics-gologging":  "telemetry/logging",
	"github_com/velonetics/velonetics-logstash":   "telemetry/logstash",
	"github_com/velonetics/velonetics-metrics":    "telemetry/metrics",
	"github_com/velonetics/velonetics-influx":     "telemetry/influx",
	"github_com/velonetics/velonetics-opencensus": "telemetry/opencensus",

	"github.com/velonetics/velonetics-lua/router":        "modifier/lua-endpoint",
	"github.com/velonetics/velonetics-lua/proxy":         "modifier/lua-proxy",
	"github.com/velonetics/velonetics-lua/proxy/backend": "modifier/lua-backend",
	"github.com/velonetics/velonetics-martian":           "modifier/martian",

}
