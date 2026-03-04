package spec

import (
	"fmt"
	"strconv"
	"strings"
)

var supported = map[string]struct{}{
	"codex":  {},
	"claude": {},
	"gemini": {},
}

func ParseAgentSpec(raw string) (map[string]int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("agent spec cannot be empty")
	}

	parts := strings.Split(trimmed, ",")
	counts := make(map[string]int)
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}
		name, countRaw, ok := strings.Cut(entry, ":")
		if !ok {
			return nil, fmt.Errorf("invalid entry %q (expected agent:count)", entry)
		}
		agent := strings.ToLower(strings.TrimSpace(name))
		if _, exists := supported[agent]; !exists {
			return nil, fmt.Errorf("unknown agent %q", agent)
		}
		count, err := strconv.Atoi(strings.TrimSpace(countRaw))
		if err != nil || count <= 0 {
			return nil, fmt.Errorf("invalid count %q for %s", countRaw, agent)
		}
		counts[agent] += count
	}
	if len(counts) == 0 {
		return nil, fmt.Errorf("agent spec cannot be empty")
	}
	return counts, nil
}
