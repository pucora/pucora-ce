package tests

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOAuth2ClientCredentialsIntegration(t *testing.T) {
	bin := gatewayBinary(t)

	const (
		clientID     = "gateway-client"
		clientSecret = "gateway-secret"
		accessToken  = "integration-test-token"
	)

	var tokenRequests atomic.Int32
	tokenServer := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		tokenRequests.Add(1)
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != clientID || pass != clientSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":"%s","expires_in":3600,"token_type":"bearer"}`, accessToken)
	})

	var backendHits atomic.Int32
	var seenAuth string
	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		backendHits.Add(1)
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer tokenServer.Close()
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
				"endpoint": "/oauth-backend",
				"backend": []map[string]interface{}{
					{
						"url_pattern": "/",
						"host":        []string{backend.URL},
						"extra_config": map[string]interface{}{
							"auth/client-credentials": map[string]interface{}{
								"client_id":     clientID,
								"client_secret": clientSecret,
								"token_url":     tokenServer.URL,
							},
						},
					},
				},
			},
		},
	}

	stop := startGateway(t, bin, cfg)
	defer stop()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/oauth-backend", gatewayPort))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", resp.StatusCode, body)
	}
	if backendHits.Load() != 1 {
		t.Fatalf("expected one backend hit, got %d", backendHits.Load())
	}
	if seenAuth != "Bearer "+accessToken {
		t.Fatalf("unexpected Authorization header: %q", seenAuth)
	}
	if tokenRequests.Load() != 1 {
		t.Fatalf("expected single token request, got %d", tokenRequests.Load())
	}
}

func TestAWSSigV4Integration(t *testing.T) {
	bin := gatewayBinary(t)

	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	t.Setenv("AWS_REGION", "us-east-1")

	var seenAuth string
	backend := startRecordingBackend(t, func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
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
			"telemetry/usage":   map[string]interface{}{"enabled": false},
		},
		"endpoints": []map[string]interface{}{
			{
				"endpoint": "/aws-backend",
				"backend": []map[string]interface{}{
					{
						"url_pattern": "/",
						"host":        []string{backend.URL},
						"extra_config": map[string]interface{}{
							"auth/aws-sigv4": map[string]interface{}{
								"service": "execute-api",
								"region":  "us-east-1",
							},
						},
					},
				},
			},
		},
	}

	stop := startGateway(t, bin, cfg)
	defer stop()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/aws-backend", gatewayPort))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", resp.StatusCode, body)
	}
	if !strings.HasPrefix(seenAuth, "AWS4-HMAC-SHA256") {
		t.Fatalf("expected SigV4 Authorization header, got %q", seenAuth)
	}
}
