package static

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
)

const clientConfigJsFileName = "client_config.js"
const gameSrcSubFolder = "game-src"

// TODO: remove this, rely only on gamepkg.ClientFolder().
const clientSubFolder = "client"

// CopyStaticResources copies all of the top-level files into the build
// directory given by dir. If copyFiles is true, it copies them, otherwise it
// symlinks them.
func CopyStaticResources(dir string, copyFiles bool) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return err
	}

	fullPkgPath, err := absoluteStaticServerPath()
	if err != nil {
		return errors.New("Couldn't get full package path: " + err.Error())
	}

	infos, err := os.ReadDir(fullPkgPath)

	if err != nil {
		return errors.New("Couldn't list files in remote directory: " + err.Error())
	}

	absLocalDirPath, err := filepath.Abs(staticDir)
	if err != nil {
		return errors.New("Couldn't resolve static build directory: " + err.Error())
	}
	absLocalDirPath, err = filepath.EvalSymlinks(absLocalDirPath)
	if err != nil {
		return errors.New("Couldn't canonicalize static build directory: " + err.Error())
	}
	absLocalDirPath += string(filepath.Separator)

	for _, info := range infos {

		name := info.Name()

		if filesToExclude[name] {
			continue
		}

		localPath := filepath.Join(staticDir, name)

		absRemotePath, err := filepath.EvalSymlinks(filepath.Join(fullPkgPath, name))
		if err != nil {
			return errors.New("Couldn't canonicalize static resource " + name + ": " + err.Error())
		}
		relRemotePath, err := path.RelativizePaths(absLocalDirPath, absRemotePath)

		if err != nil {
			return errors.New("Couldn't relativize paths: " + err.Error())
		}

		rejoinedPath := filepath.Join(absLocalDirPath, relRemotePath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {

			return errors.New("Unexpected error: relRemotePath of " + relRemotePath + " doesn't exist " + absLocalDirPath + " : " + absRemotePath + "(" + rejoinedPath + ")")
		}

		if existing, err := os.Lstat(localPath); err == nil {
			if !copyFiles || info.IsDir() {
				// It already has the requested link/directory shape.
				continue
			}
			// A dev assembly may have left a file symlink here. Production
			// copying must replace the link itself; writing through it would
			// mutate the framework source and would still leave Vite resolving
			// the input outside its build root.
			if existing.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(localPath); err != nil {
					return errors.New("Couldn't replace linked " + name + ": " + err.Error())
				}
			}
		}

		if copyFiles && !info.IsDir() {
			fmt.Println("Copying " + localPath + " to " + relRemotePath)
			if err := copyFile(absRemotePath, localPath); err != nil {
				return errors.New("Couldn't copy " + name + ": " + err.Error())
			}
		} else {
			fmt.Println("Linking " + localPath + " to " + relRemotePath)
			if err := os.Symlink(relRemotePath, localPath); err != nil {
				return errors.New("Couldn't link " + name + ": " + err.Error())
			}
		}
	}

	return nil
}

// LinkGameClientFolders creates a game-src directory and for each import listed
// in pkgs, finds a copy of that game on disk and symlinks its client directory
// into game-src.
func LinkGameClientFolders(dir string, pkgs []*gamepkg.Pkg) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return errors.New(staticDir + " doesn't exist")
	}

	gameSrcDir := filepath.Join(staticDir, gameSrcSubFolder)

	if _, err := os.Stat(gameSrcDir); os.IsNotExist(err) {
		if err := os.Mkdir(gameSrcDir, 0700); err != nil {
			return errors.New("Couldn't create game-src directory: " + err.Error())
		}
	}

	for _, pkg := range pkgs {

		absClientPath := pkg.ClientFolder()

		if absClientPath == "" {
			fmt.Println("Skipping " + pkg.Name() + " because it doesn't appear to have a client sub-directory")
			continue
		}

		relLocalPath := filepath.Join(gameSrcDir, pkg.Name())

		// Symlink targets are interpreted relative to the directory containing
		// the link, not the link path itself.
		absLocalPath, err := filepath.Abs(gameSrcDir)
		if err != nil {
			return errors.New("Couldn't resolve game-src directory: " + err.Error())
		}
		absLocalPath, err = filepath.EvalSymlinks(absLocalPath)
		if err != nil {
			return errors.New("Couldn't canonicalize game-src directory: " + err.Error())
		}
		absClientPath, err = filepath.EvalSymlinks(absClientPath)
		if err != nil {
			return errors.New("Couldn't canonicalize client directory for " + pkg.Name() + ": " + err.Error())
		}

		relPath, err := path.RelativizePaths(absLocalPath, absClientPath)

		if err != nil {
			return errors.New("Couldn't relativize path: " + err.Error())
		}

		rejoinedPath := filepath.Join(absLocalPath, relPath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {
			return errors.New("Unexpected error: relPath of " + relPath + " doesn't exist " + absLocalPath + " : " + absClientPath + "(" + rejoinedPath + ")")
		}

		if _, err := os.Stat(relLocalPath); err == nil {
			//Must already exist, so can skip
			continue
		}

		fmt.Println("Linking " + relLocalPath + " to " + relPath)
		if err := os.Symlink(relPath, relLocalPath); err != nil {
			return errors.New("Couldn't create sym lnk for " + pkg.Name() + ": " + relPath + ":: " + relLocalPath)
		}

	}

	return nil

}

// CreateClientConfigJs creates and saves a client_config.js corresponding to
// the given Clientconfig object, into the given build directory. You should use
// config.Client() to generate the ClientConfig.
func CreateClientConfigJs(dir string, c *config.ClientConfig) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return err
	}

	path := filepath.Join(staticDir, clientConfigJsFileName)

	clientBlob, err := json.MarshalIndent(c, "", "\t")

	if err != nil {
		return errors.New("Couldn't create blob: " + err.Error())
	}

	fileContents := "var CONFIG = " + string(clientBlob)

	if err := fileutil.WriteFileAtomic(path, []byte(fileContents), 0o644); err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}

	return nil

}
