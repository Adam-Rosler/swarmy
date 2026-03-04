package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
}

func writeMiniSwarmyRepo(t *testing.T, repo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, "cmd", "swarmy"), 0o755); err != nil {
		t.Fatalf("mkdir mini repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module miniswarmy\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	mainSrc := `package main
import "fmt"
func main(){ fmt.Println("mini-swarmy") }
`
	if err := os.WriteFile(filepath.Join(repo, "cmd", "swarmy", "main.go"), []byte(mainSrc), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
}

func TestInstallScriptAddsAliasLine(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash installer test is unix-only")
	}
	repoRoot := projectRoot(t)
	rcPath := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(rcPath, []byte("# test rc\n"), 0o644); err != nil {
		t.Fatalf("write rc: %v", err)
	}

	cmd := exec.Command("bash", filepath.Join(repoRoot, "scripts", "install.sh"))
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "SHELL_RC_PATH="+rcPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, string(out))
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read rc: %v", err)
	}
	if !strings.Contains(string(content), "alias swarmy=") {
		t.Fatalf("expected alias line in rc, got:\n%s", string(content))
	}
}

func TestSwarmyAutoBuildsIfMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash wrapper test is unix-only")
	}
	repoRoot := projectRoot(t)
	repo := t.TempDir()
	writeMiniSwarmyRepo(t, repo)

	cmd := exec.Command("bash", filepath.Join(repoRoot, "scripts", "swarmy-auto.sh"))
	cmd.Env = append(os.Environ(), "SWARMY_REPO_ROOT="+repo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("swarmy-auto.sh failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "mini-swarmy") {
		t.Fatalf("expected wrapped binary output, got: %s", string(out))
	}
	if _, err := os.Stat(filepath.Join(repo, "bin", "swarmy")); err != nil {
		t.Fatalf("expected built binary: %v", err)
	}
}

func TestSwarmyAutoRebuildsWhenSourceNewer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash wrapper test is unix-only")
	}
	repoRoot := projectRoot(t)
	repo := t.TempDir()
	writeMiniSwarmyRepo(t, repo)

	runner := filepath.Join(repoRoot, "scripts", "swarmy-auto.sh")
	run := func() {
		t.Helper()
		cmd := exec.Command("bash", runner)
		cmd.Env = append(os.Environ(), "SWARMY_REPO_ROOT="+repo)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("runner failed: %v\n%s", err, string(out))
		}
	}

	run()
	binPath := filepath.Join(repo, "bin", "swarmy")
	info1, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat bin after first run: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)
	mainPath := filepath.Join(repo, "cmd", "swarmy", "main.go")
	if err := os.WriteFile(mainPath, []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"mini-swarmy-v2\")}\n"), 0o644); err != nil {
		t.Fatalf("rewrite main.go: %v", err)
	}

	run()
	info2, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat bin after second run: %v", err)
	}
	if !info2.ModTime().After(info1.ModTime()) {
		t.Fatalf("expected binary to rebuild after source change: old=%v new=%v", info1.ModTime(), info2.ModTime())
	}
}
