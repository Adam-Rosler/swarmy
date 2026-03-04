package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"swarmy/internal/orchestrator"
	"swarmy/internal/preflight"
	"swarmy/internal/spec"
	"swarmy/internal/ui"
)

type Config struct {
	Task             string
	Agents           string
	CWD              string
	NoUI             bool
	JSONStream       bool
	Silent           bool
	AgentMailURL     string
	PreflightTimeout time.Duration
}

func ParseConfig(args []string) (Config, error) {
	fs := flag.NewFlagSet("swarmy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config
	fs.StringVar(&cfg.Task, "task", "", "Problem statement for swarm")
	fs.StringVar(&cfg.Agents, "agents", "", "Agent counts, e.g. codex:1,claude:2")
	fs.StringVar(&cfg.CWD, "cwd", ".", "Working directory")
	fs.BoolVar(&cfg.NoUI, "no-ui", false, "Disable Bubble Tea UI")
	fs.BoolVar(&cfg.JSONStream, "json-stream", false, "Emit newline-delimited JSON events for tool/LLM callers")
	fs.BoolVar(&cfg.Silent, "silent", false, "Non-UI summary mode for LLM/tool callers (state transitions only)")
	fs.StringVar(&cfg.AgentMailURL, "agent-mail-url", "http://127.0.0.1:8765", "Agent Mail base URL")
	timeoutSeconds := fs.Int("preflight-timeout-seconds", 2, "Preflight timeout seconds")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.Task == "" {
		return Config{}, fmt.Errorf("missing required --task")
	}
	if cfg.Agents == "" {
		return Config{}, fmt.Errorf("missing required --agents")
	}
	if cfg.NoUI && cfg.JSONStream {
		return Config{}, fmt.Errorf("--no-ui and --json-stream are mutually exclusive")
	}
	if cfg.Silent && cfg.NoUI {
		return Config{}, fmt.Errorf("--silent is already non-UI and cannot be combined with --no-ui")
	}
	cfg.CWD = filepath.Clean(cfg.CWD)
	cfg.PreflightTimeout = time.Duration(*timeoutSeconds) * time.Second
	return cfg, nil
}

func Run(args []string, stdout, stderr io.Writer) int {
	cfg, err := ParseConfig(args)
	if err != nil {
		fmt.Fprintf(stderr, "config error: %v\n", err)
		return 2
	}

	counts, err := spec.ParseAgentSpec(cfg.Agents)
	if err != nil {
		fmt.Fprintf(stderr, "spec error: %v\n", err)
		return 2
	}

	if err := runPreflight(cfg, counts); err != nil {
		fmt.Fprintf(stderr, "preflight failed: %v\n", err)
		return 1
	}

	runner := orchestrator.NewRunner(counts, cfg.Task, cfg.CWD)
	ctx := context.Background()
	events := make(chan orchestrator.Event, 256)

	if cfg.Silent {
		if cfg.JSONStream {
			return runSilentJSONStream(ctx, runner, events, stdout)
		}
		return runSilent(ctx, runner, events, stdout)
	}
	if cfg.NoUI {
		return runNoUI(ctx, runner, events, stdout)
	}
	if cfg.JSONStream {
		return runJSONStream(ctx, runner, events, stdout)
	}

	done := make(chan int, 1)
	go func() {
		code := runner.Run(ctx, events)
		done <- code
		close(done)
	}()

	program := tea.NewProgram(ui.NewLiveModel(runner.Workers, cfg.Task, events, done, false), tea.WithOutput(stdout))
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(stderr, "ui error: %v\n", err)
		return 1
	}

	final := 0
	for code := range done {
		final = code
	}
	return final
}

func runPreflight(cfg Config, counts map[string]int) error {
	if err := preflight.CheckAgentMail(cfg.AgentMailURL, cfg.PreflightTimeout); err != nil {
		return err
	}
	if err := preflight.CheckBinaries(counts, os.Getenv("PATH")); err != nil {
		return err
	}
	return nil
}

func runNoUI(ctx context.Context, runner *orchestrator.Runner, events chan orchestrator.Event, stdout io.Writer) int {
	done := make(chan int, 1)
	go func() {
		code := runner.Run(ctx, events)
		done <- code
		close(done)
	}()

	for ev := range events {
		if ev.Line != "" {
			level := ev.Level
			if level == "" {
				level = "info"
			}
			state := ev.State
			if state == "" {
				state = "-"
			}
			fmt.Fprintf(stdout, "[%s|%s|%s] %s\n", ev.AgentID, level, state, ev.Line)
		}
	}
	code := 0
	for c := range done {
		code = c
	}
	return code
}

func runJSONStream(ctx context.Context, runner *orchestrator.Runner, events chan orchestrator.Event, stdout io.Writer) int {
	done := make(chan int, 1)
	go func() {
		code := runner.Run(ctx, events)
		done <- code
		close(done)
	}()

	encoder := json.NewEncoder(stdout)
	for ev := range events {
		payload := map[string]any{
			"kind":      ev.Kind,
			"agent_id":  ev.AgentID,
			"adapter":   ev.Adapter,
			"stream":    ev.Stream,
			"level":     ev.Level,
			"state":     ev.State,
			"line":      ev.Line,
			"exit_code": ev.ExitCode,
		}
		_ = encoder.Encode(payload)
	}

	code := 0
	for c := range done {
		code = c
	}
	return code
}

func runSilent(ctx context.Context, runner *orchestrator.Runner, events chan orchestrator.Event, stdout io.Writer) int {
	done := make(chan int, 1)
	go func() {
		code := runner.Run(ctx, events)
		done <- code
		close(done)
	}()

	for ev := range events {
		if ev.Kind != "state" && ev.Kind != "done" {
			continue
		}
		state := ev.State
		if state == "" {
			state = "-"
		}
		fmt.Fprintf(stdout, "[%s] %s %s\n", ev.AgentID, ev.Kind, state)
	}
	code := 0
	for c := range done {
		code = c
	}
	return code
}

func runSilentJSONStream(ctx context.Context, runner *orchestrator.Runner, events chan orchestrator.Event, stdout io.Writer) int {
	done := make(chan int, 1)
	go func() {
		code := runner.Run(ctx, events)
		done <- code
		close(done)
	}()

	encoder := json.NewEncoder(stdout)
	for ev := range events {
		if ev.Kind != "state" && ev.Kind != "done" {
			continue
		}
		payload := map[string]any{
			"kind":      ev.Kind,
			"agent_id":  ev.AgentID,
			"adapter":   ev.Adapter,
			"state":     ev.State,
			"exit_code": ev.ExitCode,
		}
		_ = encoder.Encode(payload)
	}

	code := 0
	for c := range done {
		code = c
	}
	return code
}
