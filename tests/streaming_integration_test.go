package tests

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/encoding"
)

func TestStreamingConfigRejectedAtStartup(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*config.ServiceConfig)
		wantErr error
	}{
		{
			name: "lua post on streaming endpoint",
			mutate: func(s *config.ServiceConfig) {
				s.Endpoints[0].ExtraConfig = config.ExtraConfig{
					"github.com/pucora/pucora-lua/router": map[string]interface{}{
						"post": "local r = response.load()",
					},
				}
			},
			wantErr: config.ErrStreamingResponseManipulation,
		},
		{
			name: "backend httpcache on streaming endpoint",
			mutate: func(s *config.ServiceConfig) {
				s.Endpoints[0].Backend[0].ExtraConfig = config.ExtraConfig{
					"github.com/pucora/pucora-httpcache": map[string]interface{}{"shared": true},
				}
			},
			wantErr: config.ErrStreamingBackendHTTPCache,
		},
		{
			name: "service write_timeout with streaming endpoint",
			mutate: func(s *config.ServiceConfig) {
				s.WriteTimeout = 1
			},
			wantErr: config.ErrStreamingServiceWriteTimeout,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := streamingBaseConfig()
			tc.mutate(s)
			err := s.Init()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Init() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestStreamingConfigRejectedByCheckCommand(t *testing.T) {
	bin := "../pucora"
	if _, err := os.Stat(bin); err != nil {
		t.Skip("pucora binary not built; run make build first")
	}

	cfgPath := filepath.Join(t.TempDir(), "invalid-streaming.json")
	cfg := []byte(`{
  "version": 3,
  "port": 18080,
  "write_timeout": "30s",
  "endpoints": [{
    "endpoint": "/events",
    "output_encoding": "no-op",
    "backend": [{
      "encoding": "no-op",
      "host": ["http://127.0.0.1:8081"],
      "url_pattern": "/events"
    }]
  }]
}`)
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "check", "-c", cfgPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected check to fail, output: %s", out)
	}
	text := string(out)
	if !strings.Contains(text, "write_timeout") || !strings.Contains(strings.ToLower(text), "streaming") {
		t.Fatalf("unexpected check output: %s", out)
	}
}

func streamingBaseConfig() *config.ServiceConfig {
	return &config.ServiceConfig{
		Version: config.ConfigVersion,
		Host:    []string{"http://127.0.0.1:8081"},
		Endpoints: []*config.EndpointConfig{{
			Endpoint:       "/events",
			OutputEncoding: encoding.NOOP,
			Backend: []*config.Backend{{
				Encoding:   encoding.NOOP,
				Host:       []string{"http://127.0.0.1:8081"},
				URLPattern: "/events",
			}},
		}},
	}
}
