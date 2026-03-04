package preflight

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func CheckAgentMail(baseURL string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(baseURL)
	if err != nil {
		return fmt.Errorf("agent mail unreachable at %s: %w", baseURL, err)
	}
	_ = resp.Body.Close()
	return nil
}

func CheckBinaries(counts map[string]int, pathEnv string) error {
	var missing []string
	for agent, count := range counts {
		if count <= 0 {
			continue
		}
		if !existsOnPath(agent, pathEnv) {
			missing = append(missing, agent)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required agent binaries on PATH: %s", strings.Join(missing, ", "))
	}
	return nil
}

func existsOnPath(bin, pathEnv string) bool {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, bin)
		if isExecutable(candidate) {
			return true
		}
		if runtime.GOOS == "windows" {
			if isExecutable(candidate + ".exe") {
				return true
			}
		}
	}
	return false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	mode := info.Mode().Perm()
	return mode&0o111 != 0
}
