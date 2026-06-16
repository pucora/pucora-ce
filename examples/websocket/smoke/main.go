package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
)

func main() {
	url := flag.String("url", "", "WebSocket URL")
	message := flag.String("message", "ping", "message to send")
	header := flag.String("header", "", "optional header as Key: Value")
	expect := flag.String("expect", "", "expected exact response body")
	expectPrefix := flag.String("expect-prefix", "", "expected response prefix")
	timeout := flag.Duration("timeout", 10*time.Second, "operation timeout")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "-url is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	opts := &websocket.DialOptions{}
	if *header != "" {
		parts := strings.SplitN(*header, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid -header %q\n", *header)
			os.Exit(2)
		}
		opts.HTTPHeader = http.Header{}
		opts.HTTPHeader.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	conn, _, err := websocket.Dial(ctx, *url, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial %s: %v\n", *url, err)
		os.Exit(1)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	if err := conn.Write(ctx, websocket.MessageText, []byte(*message)); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}

	_, reply, err := conn.Read(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	got := string(reply)
	switch {
	case *expect != "" && got != *expect:
		fmt.Fprintf(os.Stderr, "expected %q, got %q\n", *expect, got)
		os.Exit(1)
	case *expectPrefix != "" && !strings.HasPrefix(got, *expectPrefix):
		fmt.Fprintf(os.Stderr, "expected prefix %q, got %q\n", *expectPrefix, got)
		os.Exit(1)
	case *expect == "" && *expectPrefix == "" && got != *message:
		fmt.Fprintf(os.Stderr, "expected echo %q, got %q\n", *message, got)
		os.Exit(1)
	}
	fmt.Printf("ok %s -> %s\n", *url, got)
}
