// Package fileutil contains safe filesystem primitives shared by
// boardgame-util commands and libraries.
package fileutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var rename = os.Rename
var link = os.Link
var removeAtomicArtifact = os.Remove

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
	cleanup := func() error {
		err := removeAtomicArtifact(tempPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove staged file %s: %w", tempPath, err)
		}
		return nil
	}

	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return errors.Join(fmt.Errorf("set permissions on temporary file for %s: %w", path, err), cleanup())
	}
	if _, err := temp.Write(contents); err != nil {
		_ = temp.Close()
		return errors.Join(fmt.Errorf("write temporary file for %s: %w", path, err), cleanup())
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return errors.Join(fmt.Errorf("sync temporary file for %s: %w", path, err), cleanup())
	}
	if err := temp.Close(); err != nil {
		return errors.Join(fmt.Errorf("close temporary file for %s: %w", path, err), cleanup())
	}
	if err := rename(tempPath, path); err != nil {
		return errors.Join(fmt.Errorf("replace %s atomically: %w", path, err), cleanup())
	}
	return nil
}

// WriteFileExclusive creates path only if it does not already exist. O_EXCL
// makes the existence check and creation one filesystem operation, avoiding a
// check-then-write race and refusing to follow an existing symlink.
func WriteFileExclusive(path string, contents []byte, mode fs.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".boardgame-exclusive-*")
	if err != nil {
		return fmt.Errorf("stage exclusive file for %s: %w", path, err)
	}
	tempPath := file.Name()
	cleanup := func() error {
		err := removeAtomicArtifact(tempPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove staged file %s: %w", tempPath, err)
		}
		return nil
	}
	if err := file.Chmod(mode.Perm()); err != nil {
		_ = file.Close()
		return errors.Join(fmt.Errorf("set permissions on exclusive file for %s: %w", path, err), cleanup())
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return errors.Join(fmt.Errorf("stage exclusive file for %s: %w", path, err), cleanup())
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return errors.Join(fmt.Errorf("sync exclusive file for %s: %w", path, err), cleanup())
	}
	if err := file.Close(); err != nil {
		return errors.Join(fmt.Errorf("close exclusive file for %s: %w", path, err), cleanup())
	}
	if err := link(tempPath, path); err != nil {
		return errors.Join(fmt.Errorf("install %s exclusively: %w", path, err), cleanup())
	}
	if err := cleanup(); err != nil {
		return fmt.Errorf("output %s committed but staging cleanup failed: %w", path, err)
	}
	return nil
}
