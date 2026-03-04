package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"swarmy/internal/cli"
)

func writeFakeAgent(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"" + name + " worker start\"\necho \"" + name + " worker done\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake agent: %v", err)
	}
}

func writeFakeAgentStderrProgress(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"" + name + " progress message\" 1>&2\necho \"" + name + " done\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake agent stderr script: %v", err)
	}
}

func TestE2EJSONStreamOutputsJSONLines(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fakes are unix-only")
	}
	binDir := t.TempDir()
	writeFakeAgent(t, binDir, "codex")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := cli.Run([]string{
		"--task", "solve it",
		"--agents", "codex:1",
		"--agent-mail-url", ts.URL,
		"--json-stream",
	}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exit, errBuf.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatalf("expected JSON output lines, got: %q", out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("expected first line to be JSON, got err: %v line=%q", err, lines[0])
	}
	if _, ok := payload["agent_id"]; !ok {
		t.Fatalf("expected agent_id key in json payload: %v", payload)
	}
}

func TestE2EJSONStreamUsesInfoForStderrProgress(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fakes are unix-only")
	}
	binDir := t.TempDir()
	writeFakeAgentStderrProgress(t, binDir, "claude")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := cli.Run([]string{
		"--task", "solve it",
		"--agents", "claude:1",
		"--agent-mail-url", ts.URL,
		"--json-stream",
	}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exit, errBuf.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	foundInfo := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("json parse failed: %v line=%q", err, line)
		}
		if payload["kind"] == "log" && payload["level"] == "info" {
			foundInfo = true
		}
	}
	if !foundInfo {
		t.Fatalf("expected stderr progress to be classified as info; output=%s", out.String())
	}
}

func TestE2ESilentOutputsStateOnlyNoLogLines(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fakes are unix-only")
	}
	binDir := t.TempDir()
	writeFakeAgent(t, binDir, "codex")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := cli.Run([]string{
		"--task", "solve it",
		"--agents", "codex:1",
		"--agent-mail-url", ts.URL,
		"--silent",
	}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exit, errBuf.String())
	}
	if strings.Contains(out.String(), "worker start") || strings.Contains(out.String(), "worker done") {
		t.Fatalf("silent mode should not emit worker log lines: %s", out.String())
	}
	if !strings.Contains(out.String(), "state") || !strings.Contains(out.String(), "done") {
		t.Fatalf("silent mode should emit state lifecycle lines: %s", out.String())
	}
}

func TestE2ESilentJSONStreamOutputsStateKindOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fakes are unix-only")
	}
	binDir := t.TempDir()
	writeFakeAgent(t, binDir, "codex")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := cli.Run([]string{
		"--task", "solve it",
		"--agents", "codex:1",
		"--agent-mail-url", ts.URL,
		"--silent",
		"--json-stream",
	}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exit, errBuf.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatalf("expected JSON output lines, got: %q", out.String())
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("expected json line, got err: %v line=%q", err, line)
		}
		kind, _ := payload["kind"].(string)
		if kind != "state" && kind != "done" {
			t.Fatalf("silent json should only emit state/done kinds, got: %v", payload)
		}
	}
}

func TestE2EPreflightFailsWhenServerDown(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := cli.Run([]string{
		"--task", "solve it",
		"--agents", "codex:1",
		"--agent-mail-url", "http://127.0.0.1:1",
		"--silent",
	}, out, errBuf)
	if exit == 0 {
		t.Fatalf("expected non-zero exit when server is down")
	}
	if !strings.Contains(errBuf.String(), "preflight") {
		t.Fatalf("expected preflight error, got: %s", errBuf.String())
	}
}
