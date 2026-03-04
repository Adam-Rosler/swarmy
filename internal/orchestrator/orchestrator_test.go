package orchestrator

import "testing"

func TestBuildWorkersFromCounts(t *testing.T) {
	workers := BuildWorkers(map[string]int{"codex": 1, "claude": 2, "gemini": 1})
	if len(workers) != 4 {
		t.Fatalf("expected 4 workers, got %d", len(workers))
	}
	want := []string{"codex-1", "claude-1", "claude-2", "gemini-1"}
	for i, id := range want {
		if workers[i].ID != id {
			t.Fatalf("worker[%d] id mismatch: got %s want %s", i, workers[i].ID, id)
		}
	}
}

func TestClassifyLineClaudeProgressOnStderrIsInfo(t *testing.T) {
	level := classifyLine("claude", "stderr", "thinking about approach")
	if level != "info" {
		t.Fatalf("expected info level, got %q", level)
	}
}

func TestClassifyLineCodexRealErrorOnStderrIsError(t *testing.T) {
	level := classifyLine("codex", "stderr", "fatal error: permission denied")
	if level != "error" {
		t.Fatalf("expected error level, got %q", level)
	}
}
