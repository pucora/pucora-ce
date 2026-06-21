package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestAPIKeysIntegration(t *testing.T) {
	bin := gatewayBinary(t)

	var backendHits atomic.Int32
	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		backendHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
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
			"telemetry/usage": map[string]interface{}{"enabled": false},
			"auth/api-keys": map[string]interface{}{
				"strategy":       "header",
				"identifier":     "Authorization",
				"keys":           []map[string]interface{}{{"key": "test-api-key-12345", "roles": []string{"user", "admin"}}},
				"propagate_role": "X-Pucora-Role",
			},
		},
		"endpoints": []map[string]interface{}{
			{
				"endpoint": "/protected",
				"backend": []map[string]interface{}{
					{"url_pattern": "/", "host": []string{backend.URL}},
				},
				"extra_config": map[string]interface{}{
					"auth/api-keys": map[string]interface{}{"roles": []string{"user"}},
				},
			},
		},
	}

	stop := startGateway(t, bin, cfg)
	defer stop()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d/protected", gatewayPort)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without key, got %d", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, baseURL, nil)
	req.Header.Set("Authorization", "Bearer test-api-key-12345")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with valid key, got %d body=%q", resp.StatusCode, body)
	}
	if backendHits.Load() != 1 {
		t.Fatalf("expected backend to be called once, got %d", backendHits.Load())
	}
}

func TestBasicAuthIntegration(t *testing.T) {
	bin := gatewayBinary(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	var backendHits atomic.Int32
	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		backendHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
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
			"telemetry/usage": map[string]interface{}{"enabled": false},
			"auth/basic": map[string]interface{}{
				"users": map[string]string{"alice": string(hash)},
			},
		},
		"endpoints": []map[string]interface{}{
			{
				"endpoint": "/secure",
				"backend": []map[string]interface{}{
					{"url_pattern": "/", "host": []string{backend.URL}},
				},
				"extra_config": map[string]interface{}{
					"auth/basic": map[string]interface{}{},
				},
			},
		},
	}

	stop := startGateway(t, bin, cfg)
	defer stop()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d/secure", gatewayPort)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without credentials, got %d", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, baseURL, nil)
	req.SetBasicAuth("alice", "secret")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with valid credentials, got %d body=%q", resp.StatusCode, body)
	}
	if backendHits.Load() != 1 {
		t.Fatalf("expected backend to be called once, got %d", backendHits.Load())
	}
}

func startRecordingBackend(t *testing.T, handler http.HandlerFunc) *httptestServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: handler}
	go srv.Serve(ln)
	t.Cleanup(func() { _ = srv.Close() })
	return &httptestServer{URL: "http://" + ln.Addr().String()}
}

type httptestServer struct {
	URL string
}

func (s *httptestServer) Close() {}

func startGateway(t *testing.T, bin string, cfg map[string]interface{}) func() {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "gateway.json")
	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(cfg); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cmd := exec.Command(bin, "run", "-c", cfgPath)
	cmd.Env = append(os.Environ(), "USAGE_DISABLE=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	port, _ := cfg["port"].(int)
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/__health", port)
	if err := waitForReady(healthURL, 15*time.Second); err != nil {
		_ = cmd.Process.Kill()
		t.Fatalf("gateway not ready: %v", err)
	}
	return func() { _ = cmd.Process.Kill() }
}
