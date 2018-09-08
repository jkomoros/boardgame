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
const configJsFileName = "client_config.js"
const gameSrcSubFolder = "game-src"
const clientSubFolder = "client"
const polymerConfig = "polymer.json"
const polymerFragmentsKey = "fragments"
const clientGameRendererFileName = "boardgame-render-game-%s.js"
const clientPlayerInfoRendererFileName = "boardgame-render-player-info-%s.js"

//The path, relative to goPath, where all of the files are to copy
const staticServerPackage = "github.com/jkomoros/boardgame/server/static"

var filesToExclude map[string]bool = map[string]bool{
	".gitignore":  true,
	"README.md":   true,
	polymerConfig: true,
	//Don't copy over because we'll generate our own; if we copy over and
	//generate our own we'll overwrite original.
	configJsFileName: true,
	".DS_Store":      true,
}

//StaticServer runs a static server. directory is the folder that the `static`
//folder is contained within. If no error is returned, runs until the program
//exits. Under the cover uses `polymer serve` because imports use bare module
//specifiers that must be rewritten.
func StaticServer(directory string, port string) error {

	if err := verifyPolymer(directory); err != nil {
		return err
	}

	staticDir := filepath.Join(directory, staticSubFolder)

	cmd := exec.Command("polymer", "serve", "--port="+port)
	cmd.Dir = staticDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.New("Couldn't `polymer serve`: " + err.Error())
	}

	return nil

}

//SimpleStaticServer creates and runs a simple static server. directory is the
//folder that the `static` folder is contained within. If no error is
//returned, Runs until the program exits.
func simpleStaticServer(directory string, port string) error {

	//This is an old implementation that came before polymer 3.0, where we really need to use polymer serve.

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

		http.Handle(name, fs)
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
//was an error. You can clean up the created folder structure with
//CleanStatic. If forceNode is true, will force update node_modules even
//if it appears to already exist. If prodBuild is true, then `polymer build`
//will be run. If copyFiles is true, instead of symlinking the files it will
//copy them (directories will still be symlinked). This is good if you intend
//to modify the files.
func Static(directory string, managers []string, c *config.Config, forceNode bool, prodBuild bool, copyFiles bool) (assetRoot string, err error) {

	if err := ensureNodeModules(forceNode); err != nil {
		return "", errors.New("node_modules couldn't be created: " + err.Error())
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

			if strings.Contains(name, "node_") {
				return "", errors.New("node_modules doesn't appear to exist. You may need to run `npm up` from within `boardgame/server/static/webapp`.")
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

		if copyFiles && !info.IsDir() {
			fmt.Println("Copying " + localPath + " to " + relRemotePath)
			if err := copyFile(absRemotePath, localPath); err != nil {
				return "", errors.New("Couldn't copy " + name + ": " + err.Error())
			}
		} else {
			fmt.Println("Linking " + localPath + " to " + relRemotePath)
			if err := os.Symlink(relRemotePath, localPath); err != nil {
				return "", errors.New("Couldn't link " + name + ": " + err.Error())
			}
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

	fmt.Println("Creating " + polymerConfig)
	if err := createPolymerJson(directory, false); err != nil {
		return "", errors.New("Couldn't create " + polymerConfig + ": " + err.Error())
	}

	if prodBuild {
		fmt.Println("Building bundled resources with `polymer build`")
		if err := buildPolymer(directory); err != nil {
			return "", errors.New("Couldn't build bundled resources: " + err.Error())
		}
	}

	return staticDir, nil

}

//copyFile copies the file at location remote to location local, copying
//cotents and perms.
func copyFile(remote, local string) error {

	info, err := os.Stat(remote)

	if err != nil {
		return errors.New("Couldn't get info for remote: " + err.Error())
	}

	contents, err := ioutil.ReadFile(remote)

	if err != nil {
		return errors.New("Couldn't read file " + remote + ": " + err.Error())
	}

	if err := ioutil.WriteFile(local, contents, info.Mode()); err != nil {
		return errors.New("Couldn't write file: " + err.Error())
	}

	return nil

}

//ensureBowerComoonents ensures that
//`$GOPATH/src/github.com/jkomoros/boardgame/server/static/webapp` has node
//components. If force is true, then will update them even if bower_components
//appears to exist.
func ensureNodeModules(force bool) error {

	p, err := path.AbsoluteGoPkgPath(staticServerPackage)

	if err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(filepath.Join(p, "node_modules")); err == nil {
			//It appears to exist, we're fine!
			return nil
		}
	}

	_, err = exec.LookPath("npm")

	if err != nil {
		return errors.New("node_modules didn't exist and npm didn't appear to be installed. You need to install npm.")
	}

	cmd := exec.Command("npm", "up")
	cmd.Dir = p
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("node_modules needs updating, running `npm up`...")
	if err := cmd.Run(); err != nil {
		return errors.New("Couldn't `npm up`: " + err.Error())
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

//createPolymerJson outputs a polymer.json in dir/static/ that includes
//fragments for all of the games in game-src. If missingFragmentsError is
//true, then if there are missing fragments we will return an error.
func createPolymerJson(dir string, missingFragmentsErrors bool) error {

	path := filepath.Join(dir, staticSubFolder, polymerConfig)

	if _, err := os.Stat(path); err == nil {
		return errors.New(path + " already exists")
	}

	fragments, missingFragments, err := listClientFragments(dir)

	if err != nil {
		return errors.New("Couldn't list fragments: " + err.Error())
	}

	if len(missingFragments) > 0 {
		if missingFragmentsErrors {
			return errors.New("The following fragments didn't exist: " + strings.Join(missingFragments, ", "))
		} else {
			for _, fragment := range missingFragments {
				fmt.Println("WARNING: missing fragment: " + fragment)
			}
		}
	}

	blob, err := polymerJsonContents(fragments)

	if err != nil {
		return errors.New("Couldn't create polymer config blob: " + err.Error())
	}

	return ioutil.WriteFile(path, blob, 0644)

}

//listClientFragments returns a list of all client fragments that should be
//merged into polymer.config given the static dir. missingFragments is
//fragments that really shouldn't have been there, but it's TECHNICALY legal
//for them to not have.
func listClientFragments(dir string) (fragments []string, missingFragments []string, err error) {

	gameSrc := filepath.Join(dir, staticSubFolder, gameSrcSubFolder)

	infos, err := ioutil.ReadDir(gameSrc)

	if err != nil {
		return nil, nil, errors.New("Couldn't list gameSrcs: " + err.Error())
	}

	for _, info := range infos {

		clientDir := filepath.Join(gameSrc, info.Name())

		gameRenderer := filepath.Join(clientDir, strings.Replace(clientGameRendererFileName, "%s", info.Name(), -1))

		if _, err := os.Stat(gameRenderer); os.IsNotExist(err) {
			missingFragments = append(missingFragments, gameRenderer)
		} else {
			fragments = append(fragments, gameRenderer)
		}

		playerInfoRenderer := filepath.Join(clientDir, strings.Replace(clientPlayerInfoRendererFileName, "%s", info.Name(), -1))

		if _, err := os.Stat(playerInfoRenderer); os.IsNotExist(err) {
			//playerInfoRenderer is optional
			continue
		}

		fragments = append(fragments, playerInfoRenderer)

	}

	for i, fragment := range fragments {
		fragments[i] = strings.TrimPrefix(fragment, filepath.Join(dir, staticSubFolder)+string(filepath.Separator))
	}

	return fragments, missingFragments, nil

}

//polymerJsonContents returns the blob representing the contents for Polymer
//json with the given fragments.
func polymerJsonContents(fragments []string) ([]byte, error) {

	var obj map[string]interface{}

	if err := json.Unmarshal([]byte(basePolymerJson), &obj); err != nil {
		return nil, errors.New("Couldn't unmarshal base json: " + err.Error())
	}

	baseFragments, ok := obj[polymerFragmentsKey]

	if !ok {
		return nil, errors.New("Base polymer json unexpectedly didn't have fragments")
	}

	interfaces, ok := baseFragments.([]interface{})

	if !ok {
		return nil, errors.New("Base polymer json's fragments unexpectedly wasn't list of interfaces")
	}

	var strs []string

	for _, inter := range interfaces {
		str, ok := inter.(string)
		if !ok {
			return nil, errors.New("One of the filename strings was unexpectedly not a string")
		}
		strs = append(strs, str)
	}

	strs = append(strs, fragments...)

	obj[polymerFragmentsKey] = strs

	return json.MarshalIndent(obj, "", "\t")

}

//verifyPolymer should be called before running a polymer command to ensure it
//will work.
func verifyPolymer(dir string) error {
	staticDir := filepath.Join(dir, staticSubFolder)

	polymerJson := filepath.Join(staticDir, polymerConfig)

	if _, err := os.Stat(polymerJson); os.IsNotExist(err) {
		return errors.New("polymer.json does not appears to exist")
	}

	_, err := exec.LookPath("polymer")

	if err != nil {
		return errors.New("polymer command is not installed. Run `npm install -g polymer-cli` to install it.")
	}

	return nil
}

//buildPolymer calls `polymer build` to build the bundled version. Expects
//polymer.json to exist; that is that createPolymerJson has been called.
func buildPolymer(dir string) error {

	staticDir := filepath.Join(dir, staticSubFolder)

	if err := verifyPolymer(dir); err != nil {
		return err
	}

	cmd := exec.Command("polymer", "build")
	cmd.Dir = staticDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.New("Couldn't `polymer build`: " + err.Error())
	}

	return nil

}

const basePolymerJson = `{
  "entrypoint": "index.html",
  "shell": "src/boardgame-app.js",
  "fragments": [
    "src/boardgame-game-view.js",
    "src/boardgame-list-games-view.js",
    "src/boardgame-404-view.js"
  ],
  "sourceGlobs": [
    "src/**/*",
    "images/**/*",
    "bower.json"
  ],
  "includeDependencies": [
    "manifest.json",
    "node_modules/@webcomponentsjs/webcomponents-loader.js"
  ]
}`
