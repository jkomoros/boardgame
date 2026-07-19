// Package fileutil contains safe filesystem primitives shared by
// boardgame-util commands and libraries.
package fileutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var rename = os.Rename

// WriteFileAtomic replaces path with contents using a same-directory temporary
// file. Existing regular-file permissions are preserved; symlinks and other
// non-regular destinations are rejected rather than followed or replaced.
func WriteFileAtomic(path string, contents []byte, defaultMode fs.FileMode) error {
	mode := defaultMode.Perm()
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to replace non-regular file %s", path)
		}
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect destination %s: %w", path, err)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".boardgame-write-*")
	if err != nil {
		return fmt.Errorf("create temporary file beside %s: %w", path, err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set permissions on temporary file for %s: %w", path, err)
	}
	if _, err := temp.Write(contents); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary file for %s: %w", path, err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary file for %s: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary file for %s: %w", path, err)
	}
	if err := rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %s atomically: %w", path, err)
	}
	return nil
}

// WriteFileExclusive creates path only if it does not already exist. O_EXCL
// makes the existence check and creation one filesystem operation, avoiding a
// check-then-write race and refusing to follow an existing symlink.
func WriteFileExclusive(path string, contents []byte, mode fs.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err != nil {
		return fmt.Errorf("create %s exclusively: %w", path, err)
	}
	complete := false
	defer func() {
		_ = file.Close()
		if !complete {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("write new file %s: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync new file %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close new file %s: %w", path, err)
	}
	complete = true
	return nil
}
