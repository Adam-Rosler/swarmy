package cli

import "testing"

func TestParseFlagsRequiresTaskAndAgents(t *testing.T) {
	if _, err := ParseConfig([]string{"--agents", "codex:1"}); err == nil {
		t.Fatal("expected missing --task error")
	}
	if _, err := ParseConfig([]string{"--task", "do thing"}); err == nil {
		t.Fatal("expected missing --agents error")
	}
}

func TestParseFlagsRejectsNoUIAndJSONStreamTogether(t *testing.T) {
	_, err := ParseConfig([]string{
		"--task", "do thing",
		"--agents", "codex:1",
		"--no-ui",
		"--json-stream",
	})
	if err == nil {
		t.Fatal("expected conflict error for --no-ui and --json-stream")
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

func TestParseFlagsRejectsSilentWithNoUI(t *testing.T) {
	if _, err := ParseConfig([]string{
		"--task", "do thing",
		"--agents", "codex:1",
		"--silent",
		"--no-ui",
	}); err == nil {
		t.Fatal("expected error for --silent with --no-ui")
	}
}
