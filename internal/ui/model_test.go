package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"swarmy/internal/orchestrator"
)

func TestModelEnterEscToggleFocus(t *testing.T) {
	m := NewModel([]orchestrator.Worker{
		{ID: "codex-1", Adapter: "codex"},
	})
	if m.FocusMode {
		t.Fatal("expected default list mode")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)
	if !m2.FocusMode {
		t.Fatal("expected focus mode after enter")
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated.(Model)
	if m3.FocusMode {
		t.Fatal("expected list mode after esc")
	}
}

func TestModelViewHasColorizedSections(t *testing.T) {
	m := NewModel([]orchestrator.Worker{{ID: "codex-1", Adapter: "codex"}})
	updated, _ := m.Update(eventMsg{event: orchestrator.Event{AgentID: "codex-1", State: "running", Line: "thinking"}, ok: true})
	view := updated.(Model).View()

	if !strings.Contains(view, "●") {
		t.Fatalf("expected status dot marker in colorized view, got: %q", view)
	}
	if !strings.Contains(view, "Swarmy") {
		t.Fatalf("expected title in view, got: %q", view)
	}
}

func TestSilentModeSuppressesLogBody(t *testing.T) {
	m := NewModel([]orchestrator.Worker{{ID: "codex-1", Adapter: "codex"}})
	m.Silent = true
	updated, _ := m.Update(eventMsg{event: orchestrator.Event{Kind: "log", AgentID: "codex-1", State: "running", Line: "private chain of thought"}, ok: true})
	m2 := updated.(Model)
	view := m2.View()

	if strings.Contains(view, "private chain of thought") {
		t.Fatalf("silent mode should hide log body, got: %q", view)
	}
	if !strings.Contains(view, "silent") {
		t.Fatalf("expected silent mode indicator, got: %q", view)
	}
}

func TestSilentModeDisablesFocusEnter(t *testing.T) {
	m := NewModel([]orchestrator.Worker{{ID: "codex-1", Adapter: "codex"}})
	m.Silent = true
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)
	if m2.FocusMode {
		t.Fatal("silent mode should keep focus mode disabled")
	}
}
