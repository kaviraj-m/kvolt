//go:build linux

package main

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestHotReloadReleasesPort(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("port/PID checks are linux-specific")
	}

	kvoltRoot := filepath.Join("..", "..") // module root from cmd/kvolt
	absRoot, err := filepath.Abs(kvoltRoot)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(absRoot, "cmd", "kvolt", "testdata", "hotreload", "main.go")
	mainPath := entry

	orig, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.WriteFile(mainPath, orig, 0644) })

	cli := "/tmp/kvolt-hotreload-test"
	if _, err := os.Stat(cli); err != nil {
		buildCLI := exec.Command("go", "build", "-o", cli, "./cmd/kvolt")
		buildCLI.Dir = absRoot
		if out, e := buildCLI.CombinedOutput(); e != nil {
			t.Fatalf("build cli: %v\n%s", e, out)
		}
	}

	cmd := exec.Command(cli, "run", "-e", entry)
	cmd.Dir = absRoot
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	waitForVersion(t, "v1", 45*time.Second)
	pid1 := listenerPID(t, "19876")
	if pid1 == "" {
		t.Fatal("no listener on :19876 after first start")
	}
	t.Logf("first server PID=%s version=v1", pid1)

	updated := strings.Replace(string(orig), `const version = "v1"`, `const version = "v2"`, 1)
	if err := os.WriteFile(mainPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid1) {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if processAlive(pid1) {
		t.Fatalf("old server PID %s still alive after reload (port leak)", pid1)
	}

	waitForVersion(t, "v2", 45*time.Second)
	pid2 := listenerPID(t, "19876")
	if pid2 == "" {
		t.Fatal("no listener on :19876 after reload")
	}
	if pid2 == pid1 {
		t.Fatalf("listener PID unchanged after reload (%s)", pid1)
	}
	t.Logf("reloaded server PID=%s version=v2", pid2)
}

func waitForVersion(t *testing.T, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:19876/version")
		if err == nil {
			buf := make([]byte, 16)
			n, _ := resp.Body.Read(buf)
			_ = resp.Body.Close()
			if strings.TrimSpace(string(buf[:n])) == want {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("server did not return version %q within %v", want, timeout)
}

func listenerPID(t *testing.T, port string) string {
	t.Helper()
	out, err := exec.Command("ss", "-tlnp").CombinedOutput()
	if err != nil {
		t.Fatalf("ss: %v", err)
	}
	needle := ":" + port
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, needle) {
			continue
		}
		if i := strings.Index(line, "pid="); i >= 0 {
			rest := line[i+4:]
			if j := strings.IndexAny(rest, ",)"); j >= 0 {
				return rest[:j]
			}
		}
	}
	return ""
}

func processAlive(pid string) bool {
	if pid == "" {
		return false
	}
	err := exec.Command("kill", "-0", pid).Run()
	return err == nil
}
