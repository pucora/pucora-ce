package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	mux.HandleFunc("/events", sseHandler)
	mux.HandleFunc("/chunked", chunkedHandler)

	log.Printf("streaming mock backend listening on %s", addr)
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

func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	count := envInt("SSE_EVENT_COUNT", 6)
	interval := envDuration("SSE_INTERVAL", 500*time.Millisecond)

	for i := 1; i <= count; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		fmt.Fprintf(w, "id: %d\n", i)
		fmt.Fprintf(w, "event: tick\n")
		fmt.Fprintf(w, "data: {\"seq\":%d,\"message\":\"event-%d\"}\n\n", i, i)
		flusher.Flush()
		time.Sleep(interval)
	}
}

func chunkedHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	chunks := []string{"alpha-", "bravo-", "charlie-"}
	for _, chunk := range chunks {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		_, _ = w.Write([]byte(chunk))
		flusher.Flush()
		time.Sleep(300 * time.Millisecond)
	}
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
