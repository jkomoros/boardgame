package main

import (
	"bytes"
	"fmt"
	"os"
)

// generatedFileCurrent compares a generated destination without reading files
// whose size already proves they differ. Missing files are stale; other
// filesystem failures and non-regular destinations are reported explicitly.
func generatedFileCurrent(path string, expected []byte) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect generated destination %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("generated destination %s is not a regular file", path)
	}
	if info.Size() != int64(len(expected)) {
		return false, nil
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read generated destination %s: %w", path, err)
	}
	return bytes.Equal(current, expected), nil
}
