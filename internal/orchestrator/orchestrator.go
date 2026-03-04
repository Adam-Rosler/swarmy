package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"swarmy/internal/adapters"
)

type Worker struct {
	ID      string
	Adapter string
	Index   int
}

type Event struct {
	Kind     string
	AgentID  string
	Adapter  string
	Stream   string
	Level    string
	State    string
	Line     string
	ExitCode int
}

type Runner struct {
	Workers []Worker
	Task    string
	CWD     string
}

func BuildWorkers(counts map[string]int) []Worker {
	workers := make([]Worker, 0)
	for _, adapter := range []string{"codex", "claude", "gemini"} {
		count := counts[adapter]
		for i := 1; i <= count; i++ {
			workers = append(workers, Worker{
				ID:      fmt.Sprintf("%s-%d", adapter, i),
				Adapter: adapter,
				Index:   i,
			})
		}
	}
	return workers
}

func NewRunner(counts map[string]int, task, cwd string) *Runner {
	return &Runner{
		Workers: BuildWorkers(counts),
		Task:    task,
		CWD:     filepath.Clean(cwd),
	}
}

func (r *Runner) Run(ctx context.Context, events chan<- Event) int {
	defer close(events)
	if len(r.Workers) == 0 {
		return 0
	}

	var wg sync.WaitGroup
	codes := make(chan int, len(r.Workers))

	for _, worker := range r.Workers {
		w := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes <- r.runWorker(ctx, w, events)
		}()
	}

	wg.Wait()
	close(codes)

	maxCode := 0
	for code := range codes {
		if code > maxCode {
			maxCode = code
		}
	}
	return maxCode
}

func (r *Runner) runWorker(ctx context.Context, worker Worker, events chan<- Event) int {
	prompt := adapters.BuildBootstrapPrompt(worker.ID, worker.Adapter, len(r.Workers), r.Task)
	cmdStr, err := adapters.BuildCommand(worker.Adapter, prompt)
	if err != nil {
		events <- Event{Kind: "done", AgentID: worker.ID, Adapter: worker.Adapter, State: "failed", Line: err.Error(), ExitCode: 1, Level: "error"}
		return 1
	}

	events <- Event{Kind: "state", AgentID: worker.ID, Adapter: worker.Adapter, State: "running", Line: "launching", Level: "info"}

	cmd := exec.CommandContext(ctx, "sh", "-lc", cmdStr)
	cmd.Dir = r.CWD

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		events <- Event{Kind: "done", AgentID: worker.ID, Adapter: worker.Adapter, State: "failed", Line: err.Error(), ExitCode: 1, Level: "error"}
		return 1
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		events <- Event{Kind: "done", AgentID: worker.ID, Adapter: worker.Adapter, State: "failed", Line: err.Error(), ExitCode: 1, Level: "error"}
		return 1
	}

	if err := cmd.Start(); err != nil {
		events <- Event{Kind: "done", AgentID: worker.ID, Adapter: worker.Adapter, State: "failed", Line: err.Error(), ExitCode: 1, Level: "error"}
		return 1
	}

	var readWG sync.WaitGroup
	readWG.Add(2)
	go func() {
		defer readWG.Done()
		scanStream(stdout, worker, "stdout", events)
	}()
	go func() {
		defer readWG.Done()
		scanStream(stderr, worker, "stderr", events)
	}()

	waitErr := cmd.Wait()
	readWG.Wait()

	code := 0
	state := "done"
	level := "info"
	if waitErr != nil {
		state = "failed"
		level = "error"
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
			if code == 0 {
				code = 1
			}
		} else {
			code = 1
		}
	}
	events <- Event{Kind: "done", AgentID: worker.ID, Adapter: worker.Adapter, State: state, Line: fmt.Sprintf("exit code %d", code), ExitCode: code, Level: level}
	return code
}

func scanStream(stream interface{ Read([]byte) (int, error) }, worker Worker, source string, events chan<- Event) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		events <- Event{
			Kind:    "log",
			AgentID: worker.ID,
			Adapter: worker.Adapter,
			Stream:  source,
			Level:   classifyLine(worker.Adapter, source, line),
			State:   "running",
			Line:    line,
		}
	}
}

func classifyLine(adapter, source, line string) string {
	if source == "stdout" {
		return "info"
	}
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return "info"
	}
	// These CLIs often emit progress and status lines on stderr; only mark as
	// errors when content strongly indicates failure.
	errMarkers := []string{
		"fatal",
		"error:",
		"exception",
		"traceback",
		"panic",
		"permission denied",
		"not found",
		"failed",
	}
	for _, marker := range errMarkers {
		if strings.Contains(lower, marker) {
			return "error"
		}
	}
	if strings.Contains(lower, "warning") || strings.Contains(lower, "warn") {
		return "warn"
	}
	_ = adapter
	return "info"
}
