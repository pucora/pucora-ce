package tests

import (
	"os"
	"testing"
)

func gatewayBinary(t *testing.T) string {
	t.Helper()
	for _, candidate := range []string{"../pucora", "../pucora.exe", "../pucora-linux"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Skip("pucora binary not built; run make build or go build -o pucora ./cmd/pucora-ce")
	return ""
}
