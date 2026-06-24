package pucora

import (
	"context"
	"strings"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/proxy"
)

const (
	conditionalNamespace = "backend/conditional"
)

type conditionalBackend struct {
	proxy     proxy.Proxy
	condition config.BackendConditional
	backend   *config.Backend
}

type conditionalRouter struct {
	logger          logging.Logger
	backends        []conditionalBackend
	fallbackBackend *conditionalBackend
}

func newConditionalRouter(logger logging.Logger, backends []*config.Backend, bf proxy.BackendFactory) proxy.Proxy {
	cr := &conditionalRouter{
		logger:          logger,
		backends:        make([]conditionalBackend, 0),
		fallbackBackend: nil,
	}

	for _, backend := range backends {
		cond := parseBackendConditional(backend.ExtraConfig)
		if cond != nil && cond.IsFallback() {
			cr.fallbackBackend = &conditionalBackend{
				proxy:     bf(backend),
				condition: *cond,
				backend:   backend,
			}
		} else if cond != nil {
			cr.backends = append(cr.backends, conditionalBackend{
				proxy:     bf(backend),
				condition: *cond,
				backend:   backend,
			})
		} else {
			cr.backends = append(cr.backends, conditionalBackend{
				proxy:     bf(backend),
				condition: config.BackendConditional{Strategy: ""},
				backend:   backend,
			})
		}
	}

	return cr.execute
}

func (cr *conditionalRouter) execute(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
	executed := false
	var lastResponse *proxy.Response
	var lastErr error

	for i := range cr.backends {
		cb := &cr.backends[i]
		if cb.condition.Strategy == "" {
			resp, err := cb.proxy(ctx, request)
			if err != nil {
				return resp, err
			}
			lastResponse = resp
			executed = true
			continue
		}

		if !cr.evaluateCondition(ctx, request, &cb.condition, cb.backend) {
			if cr.logger != nil {
				cr.logger.Debug("[BACKEND: conditional] Skipping backend due to condition not met:", cb.backend.URLPattern)
			}
			continue
		}

		resp, err := cb.proxy(ctx, request)
		if err != nil {
			return resp, err
		}
		lastResponse = resp
		executed = true
	}

	if !executed && cr.fallbackBackend != nil {
		if cr.logger != nil {
			cr.logger.Debug("[BACKEND: conditional] Executing fallback backend:", cr.fallbackBackend.backend.URLPattern)
		}
		resp, err := cr.fallbackBackend.proxy(ctx, request)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}

	return lastResponse, lastErr
}

func (cr *conditionalRouter) evaluateCondition(ctx context.Context, request *proxy.Request, cond *config.BackendConditional, backend *config.Backend) bool {
	switch cond.Strategy {
	case config.ConditionalStrategyHeader:
		return cr.evaluateHeaderStrategy(request, cond)
	case config.ConditionalStrategyPolicy:
		return cr.evaluatePolicyStrategy(ctx, request, cond, backend)
	default:
		return true
	}
}

func (cr *conditionalRouter) evaluateHeaderStrategy(request *proxy.Request, cond *config.BackendConditional) bool {
	if cond.Name == "" || cond.Value == "" {
		return false
	}

	headerValues := request.Headers[cond.Name]
	for _, v := range headerValues {
		if v == cond.Value {
			return true
		}
	}
	return false
}

func (cr *conditionalRouter) evaluatePolicyStrategy(ctx context.Context, request *proxy.Request, cond *config.BackendConditional, backend *config.Backend) bool {
	return true
}

func parseBackendConditional(extraConfig map[string]interface{}) *config.BackendConditional {
	v, ok := extraConfig[conditionalNamespace]
	if !ok || v == nil {
		return nil
	}

	cfg, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	cond := &config.BackendConditional{}
	if strategy, ok := cfg["strategy"].(string); ok {
		cond.Strategy = strategy
	}
	if name, ok := cfg["name"].(string); ok {
		cond.Name = name
	}
	if value, ok := cfg["value"].(string); ok {
		cond.Value = value
	}

	if !cond.IsValid() {
		return nil
	}

	return cond
}

func getRequiredHeaders(backends []*config.Backend) []string {
	headerSet := make(map[string]bool)
	for _, backend := range backends {
		cond := parseBackendConditional(backend.ExtraConfig)
		if cond == nil {
			continue
		}
		if cond.Strategy == config.ConditionalStrategyHeader && cond.Name != "" {
			headerSet[strings.ToLower(cond.Name)] = true
		}
	}

	headers := make([]string, 0, len(headerSet))
	for h := range headerSet {
		headers = append(headers, h)
	}
	return headers
}