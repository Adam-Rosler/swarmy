package adapters

import (
	"strings"
	"testing"
)

func TestBuildCommandIncludesYoloFlags(t *testing.T) {
	codex, err := BuildCommand("codex", "solve task")
	if err != nil {
		t.Fatal(err)
	}
	claude, err := BuildCommand("claude", "solve task")
	if err != nil {
		t.Fatal(err)
	}
	gemini, err := BuildCommand("gemini", "solve task")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(codex, "--yolo") {
		t.Fatalf("codex command missing --yolo: %s", codex)
	}
	if !strings.Contains(claude, "--dangerously-skip-permissions") {
		t.Fatalf("claude command missing skip permissions flag: %s", claude)
	}
	if !strings.Contains(gemini, "--yolo") {
		t.Fatalf("gemini command missing --yolo: %s", gemini)
	}
}

func TestBootstrapPromptNoForcedThreadOrProject(t *testing.T) {
	prompt := BuildBootstrapPrompt("codex-1", "codex", 3, "ship feature x")
	if strings.Contains(prompt, "thread_id") {
		t.Fatalf("prompt should not force thread_id: %s", prompt)
	}
	if strings.Contains(prompt, "project_key") {
		t.Fatalf("prompt should not force project_key: %s", prompt)
	}
}
