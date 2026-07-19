package static

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
)

// The main import for the main library
const mainPackage = "github.com/jkomoros/boardgame"

// The path, relative to mainPackage, where the static files are
const staticServerPath = "server/static"

const staticSubFolder = "static"

func absoluteStaticServerPath() (string, error) {

	pth, err := path.AbsoluteGoPkgPath(mainPackage)

	if err != nil {
		return "", fmt.Errorf("couldn't load main boardgame package location: %w", err)
	}

	return filepath.Join(pth, staticServerPath), nil

}

// staticBuildDir returns the static build directory within dir, creating it
// if it doesn't exist. For example, for dir="temp", returns "temp/static".
func staticBuildDir(dir string) (string, error) {
	if dir == "" {
		dir = "."
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("inspect build directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("build path %s is not a directory", dir)
	}

	staticDir := filepath.Join(dir, staticSubFolder)
	info, err = os.Lstat(staticDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0700); err != nil {
			return "", fmt.Errorf("create static directory: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("inspect static directory: %w", err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("static build path %s is not a directory", staticDir)
	}

	return staticDir, nil
}

// copyFile copies the file at location remote to location local, copying
// cotents and perms.
func copyFile(remote, local string) error {

	info, err := os.Stat(remote)

	if err != nil {
		return fmt.Errorf("inspect source file %s: %w", remote, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source path %s is not a regular file", remote)
	}

	contents, err := os.ReadFile(remote)

	if err != nil {
		return fmt.Errorf("read source file %s: %w", remote, err)
	}

	if err := fileutil.WriteFileSetAtomic(filepath.Dir(local), map[string]fileutil.FileSpec{
		filepath.Base(local): {Contents: contents, Mode: info.Mode().Perm(), ForceMode: true},
	}, true); err != nil {
		return fmt.Errorf("copy file to %s: %w", local, err)
	}

	return nil

}

// buildCachePath returns where we store our build cache (or where we WOULD if
// it existed).
func buildCachePath() (string, error) {
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", fmt.Errorf("get user cache directory: %w", err)
	}

	return filepath.Join(userCacheDir, nodeModulesCacheDir), nil
}
