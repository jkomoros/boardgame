package port

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// StaticDir is the web client's directory, relative to the module root;
// the Node helper and the engine live under it.
const StaticDir = "server/static"

// EngineDir is the TypeScript engine, relative to the module root.
const EngineDir = StaticDir + "/src/engine"

// helperScript is the Node half of port, relative to StaticDir.
const helperScript = "scripts/port-check.mjs"

// Node runs the TypeScript half of port (server/static/scripts/port-check.mjs)
// as a child process: one JSON job on stdin, one JSON result on stdout. The
// child is waited for; nothing bounds it by time.
type Node struct {
	// Binary is the node executable ("node" when empty). It must be a Node
	// that runs TypeScript (v22.18 or later; the repository uses v24).
	Binary string
	// Root is the module root.
	Root string
}

func (n Node) binary() string {
	if n.Binary == "" {
		return "node"
	}
	return n.Binary
}

// Run runs one mode of the helper with job and decodes its result into out.
func (n Node) Run(mode string, job, out any) error {
	input, err := json.Marshal(job)
	if err != nil {
		return err
	}
	cmd := exec.Command(n.binary(), helperScript, mode)
	cmd.Dir = filepath.Join(n.Root, filepath.FromSlash(StaticDir))
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("port: %s %s %s: %w\n%s", n.binary(), helperScript, mode, err, tail(stderr.String(), 4000))
	}
	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("port: %s %s gave no result: %w\n%s", helperScript, mode, err, tail(stderr.String(), 4000))
	}
	return nil
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
