package pairs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// NodeEnv names the Node binary the store's TypeScript half runs under. If
// it is unset, `node` on PATH is used.
const NodeEnv = "BOARDGAME_NODE"

// Node is the Node binary to run the TypeScript half with, or an error
// saying why none is usable. It needs Node 22.18 or later, which strips
// TypeScript types without a flag and has module.stripTypeScriptTypes.
func Node() (string, error) {
	bin := os.Getenv(NodeEnv)
	if bin == "" {
		found, err := exec.LookPath("node")
		if err != nil {
			return "", fmt.Errorf("no node on PATH; set %s to a Node 22.18+ binary", NodeEnv)
		}
		bin = found
	}
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("%s --version: %w", bin, err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if major < 22 || major == 22 && minor < 18 {
		return "", fmt.Errorf("%s is Node %s; the store's TypeScript half needs 22.18 or later (set %s)", bin, version, NodeEnv)
	}
	return bin, nil
}

// StaticDir is server/static, from the module root.
func StaticDir(root string) string {
	return filepath.Join(root, "server", "static")
}

// EngineDir is the TypeScript engine, from the module root.
func EngineDir(root string) string {
	return filepath.Join(StaticDir(root), "src", "engine")
}

// RunScript runs one of server/static/scripts with stdin and returns its
// stdout. Its stderr is returned in the error when it fails.
func RunScript(node, root, script string, stdin []byte, args ...string) ([]byte, error) {
	cmd := exec.Command(node, append([]string{"--no-warnings", filepath.Join(StaticDir(root), "scripts", script)}, args...)...)
	cmd.Dir = StaticDir(root)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w\n%s", script, strings.Join(args, " "), err, lastLines(stderr.String(), 40))
	}
	return stdout.Bytes(), nil
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
