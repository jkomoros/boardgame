package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"io/ioutil"
	"os"
	"path/filepath"
)

const staticSubFolder = "static"
const configJsFileName = "config.js"

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

	fullPkgPath, err := golden.AbsoluteGoPkgPath(staticServerPackage)

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

		relRemotePath, err := golden.RelativizePaths(absLocalDirPath, absRemotePath)

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

	return staticDir, nil

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
