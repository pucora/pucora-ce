package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	addr := envOr("LISTEN_ADDR", ":8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/CountryInfoService.wso", countryFlagHandler)

	log.Printf("mock SOAP backend listening on %s", addr)
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

func countryFlagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if action := r.Header.Get("SOAPAction"); action != "" {
		log.Printf("SOAPAction: %s", action)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bodyStr := string(body)
	if strings.Contains(bodyStr, "soapuser") && !strings.Contains(bodyStr, "soappass") {
		http.Error(w, "missing password in ws-security", http.StatusUnauthorized)
		return
	}
	code := extractCountryCode(bodyStr)
	if code == "" {
		code = "US"
	}
	flagURL := "http://www.example.com/flags/" + code + ".jpg"
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <Body>
    <CountryFlagResponse xmlns:m="http://www.example.com/countryinfo">
      <CountryFlagResult>` + flagURL + `</CountryFlagResult>
    </CountryFlagResponse>
  </Body>
</Envelope>`))
}

func extractCountryCode(body string) string {
	const open = "<sCountryISOCode>"
	const close = "</sCountryISOCode>"
	start := strings.Index(body, open)
	if start < 0 {
		return ""
	}
	start += len(open)
	end := strings.Index(body[start:], close)
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(body[start : start+end])
}
