package adapters

import (
	"fmt"
	"strconv"
	"strings"
)

var commandTemplates = map[string]string{
	"codex":  "codex --yolo exec %s",
	"claude": "claude --dangerously-skip-permissions -p %s",
	"gemini": "gemini --yolo -p %s",
}

func BuildCommand(agent, prompt string) (string, error) {
	template, ok := commandTemplates[strings.ToLower(strings.TrimSpace(agent))]
	if !ok {
		return "", fmt.Errorf("unsupported adapter %q", agent)
	}
	return fmt.Sprintf(template, strconv.Quote(prompt)), nil
}

func BuildBootstrapPrompt(workerID, adapter string, totalAgents int, task string) string {
	return fmt.Sprintf(
		"You are %s running via %s in a %d-agent swarm.\\n"+
			"MCP Agent Mail is required.\\n"+
			"Do this in order:\\n"+
			"1) Register yourself in MCP Agent Mail for the current repository context.\\n"+
			"2) Introduce yourself to all other swarm agents.\\n"+
			"3) Split work, claim file reservations before edits, and report concise progress.\\n"+
			"4) Converge on the best combined solution.\\n"+
			"Primary task:\\n%s\\n"+
			"Operate autonomously in YOLO mode.",
		workerID,
		adapter,
		totalAgents,
		task,
	)
}
