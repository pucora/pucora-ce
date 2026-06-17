package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	ws "github.com/velonetics/velonetics-websocket/v2"
)

func TestWebSocketDirectEchoIntegration(t *testing.T) {
	bin := "../velonetics"
	if _, err := os.Stat(bin); err != nil {
		t.Skip("velonetics binary not built; run make build first")
	}

	backendLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer backendLn.Close()
	backendAddr := backendLn.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !ws.IsWebSocketUpgrade(r) {
			http.Error(w, "upgrade required", http.StatusBadRequest)
			return
		}
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "bye")
		ctx := r.Context()
		for {
			typ, msg, err := c.Read(ctx)
			if err != nil {
				return
			}
			if err := c.Write(ctx, typ, msg); err != nil {
				return
			}
		}
	})
	backendSrv := &http.Server{Handler: mux}
	go backendSrv.Serve(backendLn)
	defer backendSrv.Close()

	_, backendPort, err := net.SplitHostPort(backendAddr)
	if err != nil {
		t.Fatal(err)
	}

	gatewayPort, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cfg := map[string]interface{}{
		"version": 3,
		"port":    gatewayPort,
		"extra_config": map[string]interface{}{
			"telemetry/logging": map[string]interface{}{
				"level":  "ERROR",
				"stdout": true,
			},
			"telemetry/usage": map[string]interface{}{
				"enabled": false,
			},
		},
		"endpoints": []map[string]interface{}{
			{
				"endpoint": "/ws/echo",
				"method":   "GET",
				"backend": []map[string]interface{}{
					{
						"host":                  []string{fmt.Sprintf("ws://127.0.0.1:%s", backendPort)},
						"url_pattern":           "/",
						"disable_host_sanitize": true,
					},
				},
				"extra_config": map[string]interface{}{
					"websocket": map[string]interface{}{
						"enable_direct_communication": true,
						"max_message_size":            4096,
					},
				},
			},
		},
	}

	cfgPath := filepath.Join(t.TempDir(), "ws_integration.json")
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
	defer cmd.Process.Kill()

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/__health", gatewayPort)
	if err := waitForReady(healthURL, 10*time.Second); err != nil {
		t.Fatalf("gateway not ready: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws/echo", gatewayPort)
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	const msg = "ce-ws-integration-ping"
	if err := conn.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
	typ, reply, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if typ != websocket.MessageText {
		t.Fatalf("unexpected message type: %v", typ)
	}
	if string(reply) != msg {
		t.Fatalf("unexpected reply: %q", reply)
	}
}

func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return 0, err
	}
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

func TestWebSocketFixturesReferenceValidBackendScheme(t *testing.T) {
	for _, name := range []string{"ws_direct.json", "ws_multiplex.json", "ws_jwt.json"} {
		path := filepath.Join("fixtures", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		body := string(data)
		if !strings.Contains(body, "ws://") {
			t.Fatalf("%s: expected ws:// backend host", name)
		}
		if !strings.Contains(body, `"websocket"`) {
			t.Fatalf("%s: missing websocket extra_config", name)
		}
	}
}
