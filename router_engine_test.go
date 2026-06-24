package pucora

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	luragin "github.com/pucora/lura/v2/router/gin"
)

func newTestEngine(routerOpts map[string]interface{}) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := logging.NewLogger("DEBUG", buff, "test")

	cfg := config.ServiceConfig{
		Port:         0,
		Debug:        false,
		ExtraConfig:  routerOpts,
	}

	opt := luragin.EngineOptions{
		Logger: logger,
		Writer: bytes.NewBuffer(nil),
	}

	engine := NewEngine(cfg, opt)
	return engine, func() {}
}

func TestHealthEndpoint_Enabled(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"github_com/pucora/lura/v2/router/gin": map[string]interface{}{
			"disable_health": false,
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/__health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthEndpoint_CustomPath(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"github_com/pucora/lura/v2/router/gin": map[string]interface{}{
			"health_path": "/custom/health",
			"disable_health": false,
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/custom/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthEndpoint_Disabled(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"github_com/pucora/lura/v2/router/gin": map[string]interface{}{
			"disable_health": true,
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/__health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDisableHealth_EnvVar(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"github_com/pucora/lura/v2/router/gin": map[string]interface{}{
			"disable_health": true,
			"health_path":    "/my-health",
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/my-health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestVirtualHosts_NotMatched(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"server/virtualhost": map[string]interface{}{
			"hosts": []string{"host-a.example.com", "host-b.example.com"},
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "unknown.example.com"
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestVirtualHosts_Matched(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"server/virtualhost": map[string]interface{}{
			"hosts": []string{"test.example.com"},
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "test.example.com"
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404 (no endpoint registered), got %d", w.Code)
	}
}

func TestNoHealthEndpointWithAliasHosts(t *testing.T) {
	engine, cleanup := newTestEngine(map[string]interface{}{
		"server/virtualhost": map[string]interface{}{
			"aliased_hosts": map[string]string{
				"api1": "api1.example.com",
				"api2": "api2.example.com:9000",
			},
		},
	})
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "api1.example.com"
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}