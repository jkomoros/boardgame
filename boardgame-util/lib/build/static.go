package build

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

const staticSubFolder = "static"
const configJsFileName = "config.js"
const gameSrcSubFolder = "game-src"
const clientSubFolder = "client"

//The path, relative to goPath, where all of the files are to copy
const staticServerPackage = "github.com/jkomoros/boardgame/server/static/webapp"

var filesToLink []string = []string{
	"bower.json",
	"firebase.json",
	"polymer.json",
	"manifest.json",
	"index.html",
	"src",
}

//Static creates a folder of static resources for serving within the static
//subfolder of directory. It symlinks necessary resources in. The return value
//is the directory where the assets can be served from, and an error if there
//was an error. You can clean up the created folder structure with CleanStatic.
func Static(directory string, managers []string, c *config.Config) (assetRoot string, err error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return "", errors.New(directory + " did not already exist.")
	}

	staticDir := filepath.Join(directory, staticSubFolder)

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0700); err != nil {
			return "", errors.New("Couldn't create static directory: " + err.Error())
		}
	}

	fullPkgPath, err := path.AbsoluteGoPkgPath(staticServerPackage)

	if err != nil {
		return "", errors.New("Couldn't get full package path: " + err.Error())
	}

	//TODO: some of the config files should be copied not symlinked; some of
	//these folders will stay around. Maybe take a temp parameter about
	//whehter it should do copying or not.

	workingDirectory, err := os.Getwd()

	if err != nil {
		return "", errors.New("Can't get working directory: " + err.Error())
	}

	for _, name := range filesToLink {
		localPath := filepath.Join(staticDir, name)
		absLocalDirPath := filepath.Join(workingDirectory, staticDir) + string(filepath.Separator)
		absRemotePath := filepath.Join(fullPkgPath, name)

		relRemotePath, err := path.RelativizePaths(absLocalDirPath, absRemotePath)

		rejoinedPath := filepath.Join(absLocalDirPath, relRemotePath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {
			return "", errors.New("Unexpected error: relRemotePath of " + relRemotePath + " doesn't exist " + absLocalDirPath + " : " + absRemotePath + "(" + rejoinedPath + ")")
		}

		if err != nil {
			return "", errors.New("Couldn't relativize paths: " + err.Error())
		}

		if _, err := os.Stat(localPath); err == nil {
			//Must already exist, so can skip
			continue
		}
		fmt.Println("Linking " + localPath + " to " + relRemotePath)
		if err := os.Symlink(relRemotePath, localPath); err != nil {
			return "", errors.New("Couldn't link " + name + ": " + err.Error())
		}

	}

	fmt.Println("Creating " + configJsFileName)
	if err := createConfigJs(filepath.Join(staticDir, configJsFileName), c); err != nil {
		return "", errors.New("Couldn't create " + configJsFileName + ": " + err.Error())
	}

	fmt.Println("Creating " + gameSrcSubFolder)
	if err := linkGameClientFolders(staticDir, managers); err != nil {
		return "", errors.New("Couldn't create " + gameSrcSubFolder + ": " + err.Error())
	}

	return staticDir, nil

}

//linkGameClientFolders creates a game-src within basePath and then links the
//client folders for each one.
func linkGameClientFolders(basePath string, managers []string) error {

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return errors.New(basePath + " doesn't exist")
	}

	gameSrcDir := filepath.Join(basePath, gameSrcSubFolder)

	if _, err := os.Stat(gameSrcDir); os.IsNotExist(err) {
		if err := os.Mkdir(gameSrcDir, 0700); err != nil {
			return errors.New("Couldn't create game-src directory: " + err.Error())
		}
	}

	workingDirectory, err := os.Getwd()

	if err != nil {
		return errors.New("Can't get working directory: " + err.Error())
	}

	for _, manager := range managers {
		absPkgPath, err := path.AbsoluteGoPkgPath(manager)

		if err != nil {
			return errors.New("Couldn't generate absPkgPath for " + manager + ": " + err.Error())
		}

		absClientPath := filepath.Join(absPkgPath, clientSubFolder)

		if _, err := os.Stat(absClientPath); os.IsNotExist(err) {
			fmt.Println("Skipping " + manager + " because it doesn't appear to have a client sub-directory")
			continue
		}

		pkgShortName := filepath.Base(manager)

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

func createConfigJs(path string, c *config.Config) error {
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

//CleanStatic removes all of the things created in the static subfolder within
//directory.
func CleanStatic(directory string) error {
	return os.RemoveAll(filepath.Join(directory, staticSubFolder))
}
