package fileutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fileMutation struct {
	path       string
	contents   []byte
	mode       fs.FileMode
	tempPath   string
	backupPath string
	hadFile    bool
	installed  bool
	exclusive  bool
	delete     bool
}

// FileSpec describes one file in an atomic output set.
type FileSpec struct {
	Contents []byte
	Mode     fs.FileMode
	// Delete removes the named file as part of the transaction. Missing files
	// are a no-op; directories and other non-regular paths are rejected.
	Delete bool
	// ForceMode applies Mode even when replacing an existing file. By default,
	// existing permissions are preserved.
	ForceMode bool
}

// WriteFilesAtomic validates and stages a complete set of relative files
// before replacing any destination. Installation is deterministic and rolls
// earlier replacements back if a later rename fails.
func WriteFilesAtomic(root string, files map[string][]byte, overwrite bool, defaultMode fs.FileMode) error {
	specs := make(map[string]FileSpec, len(files))
	for name, contents := range files {
		specs[name] = FileSpec{Contents: contents, Mode: defaultMode}
	}
	return WriteFileSetAtomic(root, specs, overwrite)
}

// WriteFileSetAtomic is WriteFilesAtomic with per-file creation modes.
func WriteFileSetAtomic(root string, files map[string]FileSpec, overwrite bool) error {
	if len(files) == 0 {
		return nil
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve output root: %w", err)
	}
	rootCreated := false
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return fmt.Errorf("create output root: %w", err)
		}
		rootCreated = true
		info, err = os.Stat(root)
	}
	if err != nil {
		return fmt.Errorf("inspect output root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output root %s is not a directory", root)
	}
	var createdDirs []string
	committed := false
	defer func() {
		if !committed {
			removeEmptyDirectories(createdDirs)
			if rootCreated {
				_ = os.Remove(root)
			}
		}
	}()

	mutations, err := prepareMutations(root, files, overwrite)
	if err != nil {
		return err
	}
	createdDirs, err = createDestinationDirectories(root, mutations)
	if err != nil {
		return err
	}

	if err := stageMutations(root, mutations); err != nil {
		cleanupMutationArtifacts(mutations, false)
		return err
	}
	if err := installMutations(mutations); err != nil {
		cleanupMutationArtifacts(mutations, true)
		return err
	}
	cleanupMutationArtifacts(mutations, false)
	committed = true
	return nil
}

func prepareMutations(root string, files map[string]FileSpec, overwrite bool) ([]fileMutation, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	seen := make(map[string]string, len(names))
	mutations := make([]fileMutation, 0, len(names))
	for _, name := range names {
		path, clean, err := resolveRelativeFile(root, name)
		if err != nil {
			return nil, err
		}
		if prior, ok := seen[clean]; ok {
			return nil, fmt.Errorf("output paths %q and %q resolve to the same file", prior, name)
		}
		seen[clean] = name

		spec := files[name]
		mutation := fileMutation{path: path, contents: spec.Contents, mode: spec.Mode.Perm(), exclusive: !overwrite, delete: spec.Delete}
		info, err := os.Lstat(path)
		if err == nil {
			if !info.Mode().IsRegular() {
				return nil, fmt.Errorf("refusing to replace non-regular output %s", path)
			}
			if !overwrite && !spec.Delete {
				return nil, fmt.Errorf("%s already exists; save aborted", name)
			}
			mutation.hadFile = true
			if !spec.ForceMode {
				mutation.mode = info.Mode().Perm()
			}
		} else if os.IsNotExist(err) && spec.Delete {
			continue
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect output %s: %w", path, err)
		}
		mutations = append(mutations, mutation)
	}
	return mutations, nil
}

func resolveRelativeFile(root, name string) (path, clean string, err error) {
	if name == "" || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", "", fmt.Errorf("output path %q must be a non-empty relative path", name)
	}
	clean = filepath.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("output path %q escapes the output root", name)
	}
	path = filepath.Join(root, clean)
	if err := ensureWithinRoot(root, path); err != nil {
		return "", "", fmt.Errorf("output path %q escapes the output root: %w", name, err)
	}
	return path, clean, nil
}

func ensureWithinRoot(root, path string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	ancestor := path
	for {
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			rel, err := filepath.Rel(resolvedRoot, resolved)
			if err != nil {
				return err
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("resolved path %s is outside %s", resolved, resolvedRoot)
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return fmt.Errorf("no existing ancestor for %s", path)
		}
		ancestor = parent
	}
}

func createDestinationDirectories(root string, mutations []fileMutation) ([]string, error) {
	needed := make(map[string]bool)
	for _, mutation := range mutations {
		for dir := filepath.Dir(mutation.path); dir != root; dir = filepath.Dir(dir) {
			if _, err := os.Stat(dir); err == nil {
				break
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("inspect output directory %s: %w", dir, err)
			}
			needed[dir] = true
		}
	}
	dirs := make([]string, 0, len(needed))
	for dir := range needed {
		dirs = append(dirs, dir)
	}
	sort.Slice(dirs, func(i, j int) bool {
		left, right := pathDepth(dirs[i]), pathDepth(dirs[j])
		if left != right {
			return left < right
		}
		return dirs[i] < dirs[j]
	})
	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0o755); err != nil && !os.IsExist(err) {
			removeEmptyDirectories(dirs)
			return nil, fmt.Errorf("create output directory %s: %w", dir, err)
		}
	}
	return dirs, nil
}

func stageMutations(root string, mutations []fileMutation) error {
	for i := range mutations {
		if err := ensureWithinRoot(root, mutations[i].path); err != nil {
			return fmt.Errorf("output path changed during staging: %w", err)
		}
		file, err := os.CreateTemp(filepath.Dir(mutations[i].path), ".boardgame-set-*")
		if err != nil {
			return fmt.Errorf("stage output %s: %w", mutations[i].path, err)
		}
		mutations[i].tempPath = file.Name()
		mutations[i].backupPath = file.Name() + ".backup"
		if mutations[i].delete {
			if err := file.Close(); err != nil {
				return fmt.Errorf("close deletion staging file for %s: %w", mutations[i].path, err)
			}
			if err := os.Remove(mutations[i].tempPath); err != nil {
				return fmt.Errorf("remove deletion staging file for %s: %w", mutations[i].path, err)
			}
			mutations[i].tempPath = ""
			continue
		}
		if err := file.Chmod(mutations[i].mode); err != nil {
			_ = file.Close()
			return fmt.Errorf("set staged permissions for %s: %w", mutations[i].path, err)
		}
		if _, err := file.Write(mutations[i].contents); err != nil {
			_ = file.Close()
			return fmt.Errorf("stage contents for %s: %w", mutations[i].path, err)
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			return fmt.Errorf("sync staged output for %s: %w", mutations[i].path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close staged output for %s: %w", mutations[i].path, err)
		}
	}
	return nil
}

func installMutations(mutations []fileMutation) error {
	for i := range mutations {
		if mutations[i].hadFile {
			if err := rename(mutations[i].path, mutations[i].backupPath); err != nil {
				return errors.Join(fmt.Errorf("back up %s: %w", mutations[i].path, err), rollbackMutations(mutations, i-1))
			}
		}
		if mutations[i].delete {
			mutations[i].installed = true
			continue
		}
		if mutations[i].exclusive {
			if err := link(mutations[i].tempPath, mutations[i].path); err != nil {
				return errors.Join(fmt.Errorf("install %s exclusively: %w", mutations[i].path, err), rollbackMutations(mutations, i-1))
			}
			if err := os.Remove(mutations[i].tempPath); err != nil {
				_ = os.Remove(mutations[i].path)
				return errors.Join(fmt.Errorf("remove staged link for %s: %w", mutations[i].path, err), rollbackMutations(mutations, i-1))
			}
			mutations[i].tempPath = ""
			mutations[i].installed = true
			continue
		}
		if err := rename(mutations[i].tempPath, mutations[i].path); err != nil {
			var restoreErr error
			if mutations[i].hadFile {
				restoreErr = rename(mutations[i].backupPath, mutations[i].path)
			}
			return errors.Join(fmt.Errorf("install %s: %w", mutations[i].path, err), restoreErr, rollbackMutations(mutations, i-1))
		}
		mutations[i].tempPath = ""
		mutations[i].installed = true
	}
	return nil
}

func rollbackMutations(mutations []fileMutation, last int) error {
	var errs []error
	for i := last; i >= 0; i-- {
		if !mutations[i].installed {
			continue
		}
		if err := os.Remove(mutations[i].path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove rolled-back output %s: %w", mutations[i].path, err))
			continue
		}
		mutations[i].installed = false
		if mutations[i].hadFile {
			if err := rename(mutations[i].backupPath, mutations[i].path); err != nil {
				errs = append(errs, fmt.Errorf("restore %s: %w", mutations[i].path, err))
			}
		}
	}
	return errors.Join(errs...)
}

func cleanupMutationArtifacts(mutations []fileMutation, preserveBackups bool) {
	for _, mutation := range mutations {
		if mutation.tempPath != "" {
			_ = os.Remove(mutation.tempPath)
		}
		if mutation.backupPath != "" && (!preserveBackups || !mutation.hadFile) {
			_ = os.Remove(mutation.backupPath)
		}
	}
}

func removeEmptyDirectories(dirs []string) {
	sort.Slice(dirs, func(i, j int) bool {
		left, right := pathDepth(dirs[i]), pathDepth(dirs[j])
		if left != right {
			return left > right
		}
		return dirs[i] > dirs[j]
	})
	for _, dir := range dirs {
		_ = os.Remove(dir)
	}
}

func pathDepth(path string) int {
	return strings.Count(filepath.Clean(path), string(filepath.Separator))
}
