package preflight

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckAgentMailReachable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if err := CheckAgentMail(ts.URL, 2*time.Second); err != nil {
		t.Fatalf("expected reachable server, got error: %v", err)
	}
}

func TestCheckAgentMailUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	if err := CheckAgentMail("http://"+addr, 200*time.Millisecond); err == nil {
		t.Fatal("expected unreachable server error")
	}
}

func TestCheckBinariesMissing(t *testing.T) {
	err := CheckBinaries(map[string]int{"codex": 1, "claude": 1, "gemini": 1}, "/definitely/not/a/real/path")
	if err == nil {
		t.Fatal("expected error for missing binaries")
	}
}
