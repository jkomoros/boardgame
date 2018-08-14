package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const staticSubFolder = "static"
const configJsFileName = "config.js"
const gameSrcSubFolder = "game-src"
const clientSubFolder = "client"

//The path, relative to goPath, where all of the files are to copy
const staticServerPackage = "github.com/jkomoros/boardgame/server/static/webapp"

var filesToExclude map[string]bool = map[string]bool{
	".gitignore": true,
	"README.md":  true,
	//Don't copy over because we'll generate our own; if we copy over and
	//generate our own we'll overwrite original.
	"config.js": true,
}

//SimpleStaticServer creates and runs a simple static server. directory is the
//folder that the `static` folder is contained within. If no error is
//returned, Runs until the program exits.
func SimpleStaticServer(directory string, port string) error {

	staticPath := filepath.Join(directory, staticSubFolder)

	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		return errors.New(staticPath + " does not exist")
	}

	fs := http.FileServer(http.Dir(staticPath))

	infos, err := ioutil.ReadDir(staticPath)

	if err != nil {
		return errors.New("Couldn't enumerate items in serving path")
	}

	//Install specific handlers for each existing file or directory in the
	//path to serve.
	for _, info := range infos {
		if info.Name() == "index.html" {
			continue
		}
		name := "/" + info.Name()

		if info.IsDir() {
			name += "/"
		} else {

			//Need to check if the file is a symlink to a directory, and symnlinks to directory
			//don't report as a directory in info.
			resolvedPath, err := filepath.EvalSymlinks(filepath.Join(staticPath, info.Name()))
			if err == nil {
				if info, err := os.Stat(resolvedPath); err == nil {
					if info.IsDir() {
						name += "/"
					}
				}
			}
		}

		http.Handle(name, http.StripPrefix(name, fs))
	}

	//This pattern will match as fallback (it's the shortest), and should
	//return "index.html" for everythign that doesn't match one of the ones
	//already returned.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Safe to use since "index.html" is not provided by user but is a constant
		http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
	})

	return http.ListenAndServe(":"+port, nil)

}

//Static creates a folder of static resources for serving within the static
//subfolder of directory. It symlinks necessary resources in. The return value
//is the directory where the assets can be served from, and an error if there
//was an error. You can clean up the created folder structure with CleanStatic.
func Static(directory string, managers []string, c *config.Config) (assetRoot string, err error) {

	if err := ensureBowerComponents(); err != nil {
		return "", errors.New("bower_components couldn't be created: " + err.Error())
	}

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

	infos, err := ioutil.ReadDir(fullPkgPath)

	if err != nil {
		return "", errors.New("Couldn't list files in remote directory: " + err.Error())
	}

	for _, info := range infos {

		name := info.Name()

		if filesToExclude[name] {
			continue
		}

		localPath := filepath.Join(staticDir, name)
		absLocalDirPath := filepath.Join(workingDirectory, staticDir) + string(filepath.Separator)
		absRemotePath := filepath.Join(fullPkgPath, name)

		relRemotePath, err := path.RelativizePaths(absLocalDirPath, absRemotePath)

		rejoinedPath := filepath.Join(absLocalDirPath, relRemotePath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {

			if strings.Contains(name, "bower") {
				return "", errors.New("bower_components doesn't appear to exist. You may need to run `bower update` from within `boardgame/server/static/webapp`.")
			}

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

//ensureBowerComoonents ensures that
//`$GOPATH/src/github.com/jkomoros/boardgame/server/static/webapp` has bower
//components.
func ensureBowerComponents() error {

	p, err := path.AbsoluteGoPkgPath(staticServerPackage)

	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(p, "bower_components")); err == nil {
		//It appears to exist, we're fine!
		return nil
	}

	_, err = exec.LookPath("bower")

	if err != nil {
		return errors.New("bower_components didn't exist and bower didn't appear to be installed. You need to install bower.")
	}

	cmd := exec.Command("bower", "update")
	cmd.Dir = p
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("bower_components didn't exist, running `bower update`...")
	if err := cmd.Run(); err != nil {
		return errors.New("Couldn't `bower update`: " + err.Error())
	}

	return nil

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

		pkgShortName := filepath.Base(manager)

		absClientPath := filepath.Join(absPkgPath, clientSubFolder, pkgShortName)

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
