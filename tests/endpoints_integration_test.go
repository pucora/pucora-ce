package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestPhase1Features(t *testing.T) {
	bin := "../pucora"
	if _, err := os.Stat(bin); err != nil {
		t.Skip("pucora binary not built; run make build first")
	}

	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		if r.URL.Path == "/hello" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "hello fast-json"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer backend.Close()

	t.Run("fast-json", func(t *testing.T) {
		cfgPath := "fixtures/fast-json.json"
		cfgBytes, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		// Fix backend host to point to our test server
		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/fast", gatewayPort), nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("no-redirect", func(t *testing.T) {
		cfgPath := "fixtures/no-redirect.json"
		cfgBytes, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/noredirect", gatewayPort), nil)
		// Client that doesn't follow redirects so we can check if gateway passed it back
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound {
			t.Errorf("expected 302, got %d", resp.StatusCode)
		}
		location := resp.Header.Get("Location")
		if location == "" {
			t.Errorf("expected Location header to be populated")
		}
	})
}

func TestPhase2Features(t *testing.T) {
	bin := "../pucora"
	if _, err := os.Stat(bin); err != nil {
		t.Skip("pucora binary not built; run make build first")
	}

	t.Run("jmespath", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"items": [{"id": 1, "name": "first"}, {"id": 2}]}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/jmespath.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/jmespath", gatewayPort))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), `"id":1`) {
			t.Errorf("expected response to contain \"id\":1, got: %s", body)
		}
	})

	t.Run("response-body-masking", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"card": "1234567890123456"}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/response-body.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/masked", gatewayPort))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "123456789012") {
			t.Errorf("expected original card number to be masked, but found it in response: %s", body)
		}
	})

	t.Run("request-body-extractor", func(t *testing.T) {
		var receivedHeader string
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			receivedHeader = r.Header.Get("X-User-Id")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok": true}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/request-body-extractor.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		body := strings.NewReader(`{"user_id": "u123"}`)
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/extract", gatewayPort), "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if receivedHeader != "u123" {
			t.Errorf("expected backend to receive X-User-Id=u123, got %q", receivedHeader)
		}
	})

	t.Run("security-policies-allow", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok": true}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/security-policies.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/protected", gatewayPort), nil)
		req.Header.Set("X-Api-Key", "secret")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 with valid API key, got %d", resp.StatusCode)
		}
	})

	t.Run("security-policies-deny", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok": true}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/security-policies.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/protected", gatewayPort))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 without API key, got %d", resp.StatusCode)
		}
	})

	t.Run("response-json-schema-valid", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"name": "test"}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/response-json-schema.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/validated", gatewayPort))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for valid schema response, got %d", resp.StatusCode)
		}
	})

	t.Run("response-json-schema-invalid", func(t *testing.T) {
		backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"age": 30}`))
		})
		defer backend.Close()

		cfgBytes, err := os.ReadFile("fixtures/response-json-schema.json")
		if err != nil {
			t.Fatal(err)
		}

		var cfg map[string]interface{}
		json.Unmarshal(cfgBytes, &cfg)

		endpoints := cfg["endpoints"].([]interface{})
		ep := endpoints[0].(map[string]interface{})
		backends := ep["backend"].([]interface{})
		be := backends[0].(map[string]interface{})
		be["host"] = []string{backend.URL}

		gatewayPort, err := freePort()
		if err != nil {
			t.Fatal(err)
		}
		cfg["port"] = gatewayPort

		stop := startGateway(t, bin, cfg)
		defer stop()

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/validated", gatewayPort))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500 for invalid schema response (missing required 'name'), got %d", resp.StatusCode)
		}
	})
}
