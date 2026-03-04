package spec

import "testing"

func TestParseAgentSpecValid(t *testing.T) {
	got, err := ParseAgentSpec("codex:1,claude:2,gemini:1")
	if err != nil {
		t.Fatalf("ParseAgentSpec returned error: %v", err)
	}
	want := map[string]int{"codex": 1, "claude": 2, "gemini": 1}
	if len(got) != len(want) {
		t.Fatalf("unexpected size: got %d want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("count mismatch for %s: got %d want %d", k, got[k], v)
		}
	}
}

func TestParseAgentSpecRejectsUnknown(t *testing.T) {
	_, err := ParseAgentSpec("codex:1,foo:2")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestParseAgentSpecRejectsBadCount(t *testing.T) {
	for _, tc := range []string{"codex:0", "claude:-1", "gemini:abc"} {
		if _, err := ParseAgentSpec(tc); err == nil {
			t.Fatalf("expected error for %q", tc)
		}
	}
}
