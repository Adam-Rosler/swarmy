package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"swarmy/internal/orchestrator"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	focusStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	logStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	doneStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
)

type agentState struct {
	Adapter  string
	Status   string
	Messages int
	ExitCode int
	HasExit  bool
	LastLine string
	Logs     []string
}

type Model struct {
	Workers   []orchestrator.Worker
	Order     []string
	States    map[string]*agentState
	Selected  int
	FocusMode bool
	Silent    bool
	Task      string
	Events    <-chan orchestrator.Event
	Done      <-chan int
	Finished  bool
	FinalCode int
}

type eventMsg struct {
	event orchestrator.Event
	ok    bool
}

type doneMsg struct {
	code int
	ok   bool
}

func NewModel(workers []orchestrator.Worker) Model {
	order := make([]string, 0, len(workers))
	states := make(map[string]*agentState, len(workers))
	for _, w := range workers {
		order = append(order, w.ID)
		states[w.ID] = &agentState{Adapter: w.Adapter, Status: "pending", ExitCode: -1}
	}
	return Model{Workers: workers, Order: order, States: states}
}

func NewLiveModel(workers []orchestrator.Worker, task string, events <-chan orchestrator.Event, done <-chan int, silent bool) Model {
	m := NewModel(workers)
	m.Task = task
	m.Events = events
	m.Done = done
	m.Silent = silent
	return m
}

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, 2)
	if m.Events != nil {
		cmds = append(cmds, waitEvent(m.Events))
	}
	if m.Done != nil {
		cmds = append(cmds, waitDone(m.Done))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func waitEvent(ch <-chan orchestrator.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		return eventMsg{event: ev, ok: ok}
	}
}

func waitDone(ch <-chan int) tea.Cmd {
	return func() tea.Msg {
		code, ok := <-ch
		return doneMsg{code: code, ok: ok}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "down", "j":
			if len(m.Order) > 0 {
				m.Selected = (m.Selected + 1) % len(m.Order)
			}
		case "up", "k":
			if len(m.Order) > 0 {
				m.Selected = (m.Selected - 1 + len(m.Order)) % len(m.Order)
			}
		case "enter":
			if !m.Silent {
				m.FocusMode = true
			}
		case "esc":
			m.FocusMode = false
		}
	case eventMsg:
		if !msg.ok {
			return m, nil
		}
		st := m.States[msg.event.AgentID]
		if st != nil {
			if msg.event.State != "" {
				st.Status = msg.event.State
			}
			if msg.event.Line != "" {
				st.Messages++
				if m.Silent {
					st.LastLine = "silent mode: log body hidden"
				} else {
					line := formatEventLine(msg.event)
					st.LastLine = line
					st.Logs = append(st.Logs, line)
					if len(st.Logs) > 500 {
						st.Logs = st.Logs[len(st.Logs)-500:]
					}
				}
			}
			if msg.event.Kind == "done" {
				st.ExitCode = msg.event.ExitCode
				st.HasExit = true
			}
		}
		return m, waitEvent(m.Events)
	case doneMsg:
		if msg.ok {
			m.Finished = true
			m.FinalCode = msg.code
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	if m.FocusMode {
		return m.focusView()
	}
	return m.listView()
}

func (m Model) listView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Swarmy (Bubble Tea)") + "\n")
	if m.Task != "" {
		b.WriteString(subtleStyle.Render("Task: "+m.Task) + "\n")
	}
	keys := "Keys: j/k move, enter focus, esc back, q quit"
	if m.Silent {
		keys = "Keys: j/k move, q quit"
	}
	b.WriteString(subtleStyle.Render(keys) + "\n")
	if m.Silent {
		b.WriteString(subtleStyle.Render("silent mode enabled: live log text is hidden") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Idx  Agent      Adapter  State       Msgs  Exit  Last") + "\n")
	b.WriteString(subtleStyle.Render("---  ---------  -------  ----------  ----  ----  ------------------------------") + "\n")
	for i, id := range m.Order {
		st := m.States[id]
		if st == nil {
			continue
		}
		pointer := " "
		if i == m.Selected {
			pointer = ">"
		}
		exit := "-"
		if st != nil && st.HasExit {
			exit = fmt.Sprintf("%d", st.ExitCode)
		}
		last := ""
		if st != nil {
			last = truncate(st.LastLine, 30)
		}
		state := "● " + st.Status
		rowText := fmt.Sprintf("%s%-3d %-9s  %-7s  %-10s  %-4d  %-4s  %s",
			pointer,
			i+1,
			id,
			st.Adapter,
			state,
			st.Messages,
			exit,
			last,
		)
		if i == m.Selected {
			b.WriteString(selectedStyle.Render(rowText) + "\n")
		} else {
			b.WriteString(rowText + "\n")
		}
	}
	if len(m.Order) > 0 {
		selectedID := m.Order[m.Selected]
		st := m.States[selectedID]
		b.WriteString("\n" + headerStyle.Render("Selected: "+selectedID) + "\n")
		if m.Silent {
			b.WriteString(subtleStyle.Render("silent mode: drill-down log view disabled") + "\n")
		} else {
			for _, line := range tail(st.Logs, 20) {
				b.WriteString(logStyle.Render(line) + "\n")
			}
		}
	}
	if m.Finished {
		b.WriteString("\n" + doneStyle.Render(fmt.Sprintf("Run complete. Exit code: %d", m.FinalCode)) + "\n")
	}
	return b.String()
}

func (m Model) focusView() string {
	if len(m.Order) == 0 {
		return "No workers."
	}
	id := m.Order[m.Selected]
	st := m.States[id]
	var b strings.Builder
	b.WriteString(focusStyle.Render(fmt.Sprintf("Agent Focus: %s (%s)", id, st.Adapter)) + "\n")
	b.WriteString(subtleStyle.Render("Keys: esc back, j/k move selection, q quit") + "\n")
	b.WriteString(subtleStyle.Render(fmt.Sprintf("State: ● %s  Messages: %d", st.Status, st.Messages)) + "\n\n")
	for _, line := range tail(st.Logs, 200) {
		b.WriteString(logStyle.Render(line) + "\n")
	}
	if len(st.Logs) == 0 {
		b.WriteString(subtleStyle.Render("No output yet.") + "\n")
	}
	return b.String()
}

func tail(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	return lines[len(lines)-n:]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func formatEventLine(ev orchestrator.Event) string {
	level := ev.Level
	if level == "" {
		level = "info"
	}
	if ev.Kind == "log" {
		return fmt.Sprintf("[%s] %s", level, ev.Line)
	}
	return ev.Line
}
