# swarmy

`swarmy` is a Go CLI that launches mixed-agent swarms (`codex`, `claude`, `gemini`) in **YOLO mode** and coordinates progress with MCP Agent Mail.

Each spawned worker runs with YOLO permissions:
- `codex --yolo ...`
- `claude --dangerously-skip-permissions ...`
- `gemini --yolo ...`

## Requirements

- Go (1.25+ recommended)
- MCP Agent Mail server running (default: `http://127.0.0.1:8765`)
- Agent CLIs on `PATH`: `codex`, `claude`, `gemini`

## Build

```bash
cd /Users/adamrosler/Documents/Code/swarmy
go mod tidy
go build -o bin/swarmy ./cmd/swarmy
```

## Rebuild

Use this anytime you change code:

```bash
cd /Users/adamrosler/Documents/Code/swarmy
go build -o bin/swarmy ./cmd/swarmy
```

## Run Modes

### 1) Interactive Bubble Tea UI (default)

```bash
./bin/swarmy \
  --task "Implement feature X and ship tests" \
  --agents "codex:1,claude:2,gemini:1"
```

### 2) Plain log mode (`--no-ui`)

Streams parsed log levels (`info`, `warn`, `error`) with state:

```bash
./bin/swarmy \
  --task "Implement feature X and ship tests" \
  --agents "codex:1,claude:2,gemini:1" \
  --no-ui
```

### 3) Full machine-readable event stream (`--json-stream`)

Newline-delimited JSON (all event types):

```bash
./bin/swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --json-stream
```

### 4) Low-token lifecycle mode (`--silent`)

No log body; only state/done lifecycle updates:

```bash
./bin/swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --silent
```

### 5) Silent JSON mode (`--silent --json-stream`)

State-only NDJSON (best for LLM/tool orchestration loops):

```bash
./bin/swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --silent \
  --json-stream
```

## Flags

- `--task` required
- `--agents` required, format: `codex:1,claude:2,gemini:1`
- `--no-ui` optional
- `--json-stream` optional
- `--silent` optional (`--silent` cannot be combined with `--no-ui`)
- `--cwd` optional (default `.`)
- `--agent-mail-url` optional (default `http://127.0.0.1:8765`)
- `--preflight-timeout-seconds` optional (default `2`)

## Tell Your Agent To Build + Run

Use one of these prompts directly with your coding agent.

### Generic instruction

```text
In /Users/adamrosler/Documents/Code/swarmy:
1) Build: go build -o bin/swarmy ./cmd/swarmy
2) Run: ./bin/swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
3) If build fails, fix and rebuild.
4) If preflight fails, report exact error and stop.
```

### Codex prompt

```text
Build and run swarmy from /Users/adamrosler/Documents/Code/swarmy.
Use: go build -o bin/swarmy ./cmd/swarmy
Then run:
./bin/swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
Return only parsed JSON lifecycle events and final exit code.
```

### Claude/Gemini prompt

```text
From /Users/adamrosler/Documents/Code/swarmy, compile and execute swarmy.
Build: go build -o bin/swarmy ./cmd/swarmy
Run:
./bin/swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
If MCP preflight fails, include the exact failure message.
```

## Tests

```bash
go test ./...
./scripts/test-unit.sh
./scripts/test-e2e.sh
./scripts/test-all.sh
```

## Related

- MCP Agent Mail by Jeffrey (Dicklesworthstone): https://github.com/Dicklesworthstone/mcp_agent_mail
