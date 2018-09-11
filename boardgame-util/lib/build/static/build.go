package static

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"io/ioutil"
	"os"
	"path/filepath"
)

const clientConfigJsFileName = "client_config.js"
const gameSrcSubFolder = "game-src"
const clientSubFolder = "client"

const clientGameRendererFileName = "boardgame-render-game-%s.js"
const clientPlayerInfoRendererFileName = "boardgame-render-player-info-%s.js"

//CopyStaticResources copies all of the top-level files into the build
//directory given by dir. If copyFiles is true, it copies them, otherwise it
//symlinks them.
func CopyStaticResources(dir string, copyFiles bool) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return err
	}

	fullPkgPath, err := absoluteStaticServerPath()
	if err != nil {
		return errors.New("Couldn't get full package path: " + err.Error())
	}

	workingDirectory, err := os.Getwd()

	if err != nil {
		return errors.New("Can't get working directory: " + err.Error())
	}

	infos, err := ioutil.ReadDir(fullPkgPath)

	if err != nil {
		return errors.New("Couldn't list files in remote directory: " + err.Error())
	}

	absLocalDirPath := filepath.Join(workingDirectory, staticDir) + string(filepath.Separator)

	for _, info := range infos {

		name := info.Name()

		if filesToExclude[name] {
			continue
		}

		localPath := filepath.Join(staticDir, name)

		absRemotePath := filepath.Join(fullPkgPath, name)
		relRemotePath, err := path.RelativizePaths(absLocalDirPath, absRemotePath)

		if err != nil {
			return errors.New("Couldn't relativize paths: " + err.Error())
		}

		rejoinedPath := filepath.Join(absLocalDirPath, relRemotePath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {

			return errors.New("Unexpected error: relRemotePath of " + relRemotePath + " doesn't exist " + absLocalDirPath + " : " + absRemotePath + "(" + rejoinedPath + ")")
		}

		if _, err := os.Stat(localPath); err == nil {
			//Must already exist, so can skip
			continue
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

//LinkGameClientFoldrs creates a game-src directory and for each import listed
//in gameImports, finds a copy of that game on disk and symlinks its client
//directory into game-src.
func LinkGameClientFolders(dir string, gameImports []string) error {

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

	workingDirectory, err := os.Getwd()

	if err != nil {
		return errors.New("Can't get working directory: " + err.Error())
	}

	for _, manager := range gameImports {
		absPkgPath, err := path.AbsoluteGoPkgPath(manager)

		if err != nil {
			return errors.New(manager + " didn't seem to be installed or installable: " + err.Error())
		}

		pkgShortName := filepath.Base(manager)

		absClientPath := filepath.Join(absPkgPath, clientSubFolder)

		if _, err := os.Stat(absClientPath); os.IsNotExist(err) {
			fmt.Println("Skipping " + manager + " because it doesn't appear to have a client sub-directory")
			continue
		}

		relLocalPath := filepath.Join(gameSrcDir, pkgShortName)

		//This feels like it should be relLocalPath, but it needs to be
		//gameSrcDir, otherwise there's an extra ".." in the path. Not really
		//sure why. :-/
		absLocalPath := filepath.Join(workingDirectory, gameSrcDir)

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
			return errors.New("Couldn't create sym lnk for " + manager + ": " + relPath + ":: " + relLocalPath)
		}

	}

	return nil

}

//CreateClientConfigJs creates and saves a client_config.js corresponding to
//the given top-level config object, into the given build directory. It uses
//config.Client() to generate the contents for the file.
func CreateClientConfigJs(dir string, c *config.Config) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return err
	}

	path := filepath.Join(staticDir, clientConfigJsFileName)

	client := c.Client(false)

	clientBlob, err := json.MarshalIndent(client, "", "\t")

	if err != nil {
		return errors.New("Couldn't create blob: " + err.Error())
	}

	fileContents := "var CONFIG = " + string(clientBlob)

	if err := ioutil.WriteFile(path, []byte(fileContents), 0644); err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}

	return nil

}
