package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

const hs256KeyB64 = "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}
	addr := envOr("LISTEN_ADDR", ":8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/jwk/symmetric", symmetricJWKEndpoint)
	mux.HandleFunc("/token", tokenHandler)
	log.Printf("mock jwt listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
		"token": token,
		"value": "Bearer " + token,
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
	}
	priv := struct {
		Roles []string `json:"roles"`
	}{Roles: []string{"role_a", "role_b"}}
	return jwt.Signed(signer).Claims(claims).Claims(priv).CompactSerialize()
}
