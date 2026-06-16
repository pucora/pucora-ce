package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

const (
	handshakeMessage = `{"msg":"Velonetics WS proxy starting"}`
	handshakeOK      = "OK"
	hs256KeyB64      = "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	addr := envOr("LISTEN_ADDR", ":8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/jwk/symmetric", symmetricJWKEndpoint)
	mux.HandleFunc("/token", tokenHandler)
	mux.HandleFunc("/ws", multiplexHandler)
	mux.HandleFunc("/echo", directEchoHandler)

	log.Printf("mock backend listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func runHealthcheck() {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8081/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

func symmetricJWKEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{
  "keys": [
    {
      "kty": "oct",
      "alg": "A128KW",
      "k": "GawgguFyGrWKav7AX4VKUg",
      "kid": "sim1"
    },
    {
      "kty": "oct",
      "k": "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow",
      "kid": "sim2",
      "alg": "HS256"
    }
  ]
}`))
}

func tokenHandler(w http.ResponseWriter, _ *http.Request) {
	token, err := signTestToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token":      token,
		"header":     "Authorization",
		"value":      "Bearer " + token,
		"expires_in": "24h",
	})
}

func signTestToken() (string, error) {
	key, err := base64.RawURLEncoding.DecodeString(hs256KeyB64)
	if err != nil {
		return "", err
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "sim2"),
	)
	if err != nil {
		return "", err
	}
	claims := jwt.Claims{
		Issuer:   "http://example.com",
		Audience: jwt.Audience{"http://api.example.com"},
		Expiry:   jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		ID:       "compose-test-jti",
	}
	priv := struct {
		Sub   string   `json:"sub"`
		Roles []string `json:"roles"`
	}{
		Sub:   "compose-user",
		Roles: []string{"role_a", "role_b"},
	}
	return jwt.Signed(signer).Claims(claims).Claims(priv).CompactSerialize()
}

func directEchoHandler(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "websocket upgrade required", http.StatusBadRequest)
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
}

func multiplexHandler(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "websocket upgrade required", http.StatusBadRequest)
		return
	}
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "bye")
	ctx := r.Context()

	_, data, err := c.Read(ctx)
	if err != nil {
		return
	}
	if string(data) != handshakeMessage {
		log.Printf("unexpected handshake: %s", string(data))
		return
	}
	if err := c.Write(ctx, websocket.MessageText, []byte(handshakeOK)); err != nil {
		return
	}

	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		reply, err := multiplexReply(data)
		if err != nil {
			if err.Error() == "skip empty envelope" {
				continue
			}
			log.Printf("multiplex reply: %v", err)
			continue
		}
		if err := c.Write(ctx, websocket.MessageText, reply); err != nil {
			return
		}
	}
}

type envelope struct {
	URL     string                 `json:"url"`
	Session map[string]interface{} `json:"session"`
	Body    string                 `json:"body"`
}

func multiplexReply(data []byte) ([]byte, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	payload, err := base64.StdEncoding.DecodeString(env.Body)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(payload))
	if text == "" {
		return nil, fmt.Errorf("skip empty envelope")
	}
	text = "echo:" + text
	out := envelope{
		URL:     env.URL,
		Session: env.Session,
		Body:    base64.StdEncoding.EncodeToString([]byte(text)),
	}
	return json.Marshal(out)
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Connection"), "upgrade") &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}
