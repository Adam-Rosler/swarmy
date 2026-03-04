package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseFlagsRequiresTaskAndAgents(t *testing.T) {
	if _, err := ParseConfig([]string{"--agents", "codex:1"}); err == nil {
		t.Fatal("expected missing --task error")
	}
	if _, err := ParseConfig([]string{"--task", "do thing"}); err == nil {
		t.Fatal("expected missing --agents error")
	}
}

func TestParseFlagsRejectsRemovedNoUIFlag(t *testing.T) {
	_, err := ParseConfig([]string{
		"--task", "do thing",
		"--agents", "codex:1",
		"--no-ui",
	})
	if err == nil {
		t.Fatal("expected unknown/invalid flag error for removed --no-ui")
	}
}

func TestParseFlagsAcceptsSilent(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"--task", "do thing",
		"--agents", "codex:1",
		"--silent",
	})
	if err != nil {
		t.Fatalf("expected no error for --silent, got: %v", err)
	}
	if !cfg.Silent {
		t.Fatal("expected cfg.Silent true")
	}
}

func TestParseFlagsAllowsSilentWithJSONStream(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"--task", "do thing",
		"--agents", "codex:1",
		"--silent",
		"--json-stream",
	})
	if err != nil {
		t.Fatalf("expected no error for --silent with --json-stream, got: %v", err)
	}
	if !cfg.Silent || !cfg.JSONStream {
		t.Fatal("expected silent and json_stream enabled")
	}
}

func TestRunHelpFlagPrintsHelpAndReturnsZero(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := Run([]string{"--help"}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0 for help, got %d", exit)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help usage in stdout, got: %s", out.String())
	}
}

func TestRunShortHelpFlagPrintsHelpAndReturnsZero(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	exit := Run([]string{"-h"}, out, errBuf)
	if exit != 0 {
		t.Fatalf("expected exit 0 for -h, got %d", exit)
	}
	if !strings.Contains(out.String(), "--silent") {
		t.Fatalf("expected --silent in help output, got: %s", out.String())
	}
}

func TestHelpTextIncludesSilentAndJSONFlags(t *testing.T) {
	help := HelpText()
	if !strings.Contains(help, "--silent") {
		t.Fatalf("help missing --silent: %s", help)
	}
	if !strings.Contains(help, "--json-stream") {
		t.Fatalf("help missing --json-stream: %s", help)
	}
	if !strings.Contains(help, "preflight") {
		t.Fatalf("help should explain preflight timeout meaning: %s", help)
	}
}
