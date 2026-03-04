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

## Install Alias (Auto-Build + Auto-Rebuild)

This project includes an installer that:
- builds `bin/swarmy`,
- adds `alias swarmy='.../scripts/swarmy-auto.sh'` to your `~/.bashrc`,
- ensures every `swarmy` invocation rebuilds automatically if Go source files changed.
- supports overriding the rc file via `SHELL_RC_PATH=/path/to/rc` when needed.

Install:

```bash
cd /Users/adamrosler/Documents/Code/swarmy
./scripts/install.sh
source ~/.bashrc
```

After that, just run:

```bash
swarmy --task "..." --agents "codex:1,claude:1,gemini:1"
```

What the alias does:
- points `swarmy` to `scripts/swarmy-auto.sh`,
- auto-builds if `bin/swarmy` is missing,
- auto-rebuilds when any `*.go`, `go.mod`, or `go.sum` file is newer than the binary.

## Run Modes

Commands below assume you used `./scripts/install.sh` and sourced your shell rc so `swarmy` alias is available.
If you did not install the alias, replace `swarmy` with `./bin/swarmy`.

### 1) Interactive Bubble Tea UI (default)

```bash
swarmy \
  --task "Implement feature X and ship tests" \
  --agents "codex:1,claude:2,gemini:1"
```

### 2) Full machine-readable event stream (`--json-stream`)

Newline-delimited JSON (all event types):

```bash
swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --json-stream
```

### 3) Low-token lifecycle mode (`--silent`)

No log body; only state/done lifecycle updates:

```bash
swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --silent
```

### 4) Silent JSON mode (`--silent --json-stream`)

State-only NDJSON (best for LLM/tool orchestration loops):

```bash
swarmy \
  --task "Solve X" \
  --agents "codex:1,claude:1" \
  --silent \
  --json-stream
```

## Safety / Token Cost

Running swarms can burn a lot of tokens quickly because multiple agents run at once in YOLO mode.

Use with care:
- Start with small counts (`codex:1,claude:1`) before scaling up.
- Prefer `--silent --json-stream` for orchestration loops to reduce output noise.
- Reserve bigger swarms for high-value tasks.

## Flags

- `--task` required
- `--agents` required, format: `codex:1,claude:2,gemini:1`
- `--json-stream` optional
- `--silent` optional
- `--cwd` optional (default `.`)
- `--agent-mail-url` optional (default `http://127.0.0.1:8765`), used for MCP Agent Mail preflight checks
- `--preflight-timeout-seconds` optional (default `2`), timeout for that MCP Agent Mail preflight connectivity check
- `-h`, `--help` show usage/help menu

## Preflight Timeout

`--preflight-timeout-seconds` controls how long `swarmy` waits while checking MCP Agent Mail is reachable **before spawning any agents**.
If the check times out or fails, the run exits immediately.

## Help Menu

```bash
swarmy --help
```

## Tell Your Agent How To Use Swarmy (Alias Installed)

Use these snippets when `./scripts/install.sh` has already been run and `swarmy` alias is available in the shell.

### Generic instruction

```text
In /Users/adamrosler/Documents/Code/swarmy:
1) Run: swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
2) Parse lifecycle JSON events.
3) Report final exit code and any preflight failures verbatim.
```

### Codex prompt

```text
Run swarmy from /Users/adamrosler/Documents/Code/swarmy using the installed alias:
swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
Return only parsed JSON lifecycle events and final exit code.
```

### Claude/Gemini prompt

```text
From /Users/adamrosler/Documents/Code/swarmy, execute:
swarmy --task "<TASK>" --agents "codex:1,claude:2,gemini:1" --silent --json-stream
If MCP preflight fails, include the exact failure message.
```

## Agent Rules Snippet

You can add this to your agent rules/instructions:

```text
When using swarmy:
- Assume alias is already installed via scripts/install.sh
- Use conservative defaults to avoid token burn.
- Prefer: swarmy --task "<TASK>" --agents "codex:1,claude:1" --silent --json-stream
- Parse lifecycle JSON events and report final exit code.
- If preflight fails, stop and return the exact error.
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
