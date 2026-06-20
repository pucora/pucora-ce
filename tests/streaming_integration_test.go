package tests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/encoding"
)

func TestStreamingRuntimeIntegration(t *testing.T) {
	bin := "../pucora"
	if _, err := os.Stat(bin); err != nil {
		t.Skip("pucora binary not built; run make build first")
	}

	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.ResponseWriter to be an http.Flusher")
		}

		_, _ = w.Write([]byte("data: chunk1\n\n"))
		flusher.Flush()

		if strings.HasPrefix(r.URL.Path, "/fail") {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
			return
		}

		_, _ = w.Write([]byte("data: chunk2\n\n"))
		flusher.Flush()
	})
	defer backend.Close()

	gatewayPort, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cfg := map[string]interface{}{
		"version": 3,
		"port":    gatewayPort,
		"extra_config": map[string]interface{}{
			"telemetry/logging": map[string]interface{}{"level": "ERROR", "stdout": true},
			"telemetry/usage":   map[string]interface{}{"enabled": false},
		},
		"endpoints": []map[string]interface{}{
			{
				"endpoint":        "/stream",
				"output_encoding": "no-op",
				"backend": []map[string]interface{}{
					{
						"url_pattern": "/stream",
						"host":        []string{backend.URL},
						"encoding":    "no-op",
					},
				},
			},
			{
				"endpoint":        "/fail",
				"output_encoding": "no-op",
				"backend": []map[string]interface{}{
					{
						"url_pattern": "/fail",
						"host":        []string{backend.URL},
						"encoding":    "no-op",
					},
				},
			},
		},
	}

	stop := startGateway(t, bin, cfg)
	defer stop()

	t.Run("sse-chunked", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/stream", gatewayPort), nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Type") != "text/event-stream" {
			t.Errorf("expected Content-Type text/event-stream, got %q", resp.Header.Get("Content-Type"))
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		res := string(b)
		if !strings.Contains(res, "data: chunk1") {
			t.Errorf("expected chunk1, got %q", res)
		}
		if !strings.Contains(res, "data: chunk2") {
			t.Errorf("expected chunk2, got %q", res)
		}
	})

	t.Run("mid-stream-failure", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/fail", gatewayPort), nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		buf := make([]byte, 1024)
		n, err := resp.Body.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		res := string(buf[:n])
		if !strings.Contains(res, "data: chunk1") {
			t.Errorf("expected chunk1, got %q", res)
		}

		_, err = resp.Body.Read(buf)
		if err == nil {
			t.Fatal("expected error on mid-stream failure, got nil")
		}
	})
}

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
