package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"io/ioutil"
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
const packageJsonFileName = "package.json"
const clientGameRendererFileName = "boardgame-render-game-%s.js"
const clientPlayerInfoRendererFileName = "boardgame-render-player-info-%s.js"
const nodeModulesFolder = "node_modules"

//The main import for the main library
const mainPackage = "github.com/jkomoros/boardgame"

//The path, relative to mainPackage, where the static files are
const staticServerPath = "server/static"

//The name of the direcotry within os.UserCacheDir() that node_modules should
//be created within.
const nodeModulesCacheDir = "com.github.jkomoros.boardgame"

var filesToExclude map[string]bool = map[string]bool{
	".gitignore":      true,
	"README.md":       true,
	polymerConfig:     true,
	nodeModulesFolder: true,
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

//CleanCache clears the central cache the build system uses (currently just
//node_modules). If that cache doesn't exist, is a no op.
func CleanCache() error {

	cacheDir, err := buildCachePath()
	if err != nil {
		return errors.New("Couldn't get build cache path: " + err.Error())
	}

	//os.RemoveAll is OK if the path doesn't exist
	return os.RemoveAll(cacheDir)

}

//StaticBuildDir returns the static build directory within dir, creating it
//if it doesn't exist. For example, for dir="temp", returns "temp/static".
func StaticBuildDir(dir string) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", errors.New(dir + " did not already exist.")
	}

	staticDir := filepath.Join(dir, staticSubFolder)

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0700); err != nil {
			return "", errors.New("Couldn't create static directory: " + err.Error())
		}
	}

	return staticDir, nil
}

//Static creates a folder of static resources for serving within the static
//subfolder of directory. It symlinks necessary resources in. The return value
//is the directory where the assets can be served from, and an error if there
//was an error. You can clean up the created folder structure with
//CleanStatic. If prodBuild is true, then `polymer build` will be run. If
//copyFiles is true, instead of symlinking the files it will copy them
//(directories will still be symlinked). This is good if you intend to modify
//the files.
func Static(directory string, managers []string, c *config.Config, prodBuild bool, copyFiles bool) (assetRoot string, err error) {

	staticDir, err := StaticBuildDir(directory)
	if err != nil {
		return "", err
	}

	fullPkgPath, err := absoluteStaticServerPath()
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
			return "", errors.New("Couldn't relativize paths: " + err.Error())
		}

		rejoinedPath := filepath.Join(absLocalDirPath, relRemotePath)

		if _, err := os.Stat(rejoinedPath); os.IsNotExist(err) {

			return "", errors.New("Unexpected error: relRemotePath of " + relRemotePath + " doesn't exist " + absLocalDirPath + " : " + absRemotePath + "(" + rejoinedPath + ")")
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

	//Ensure node_modules exists adn link to it
	absRemoteNodePath, err := updateNodeModules(filepath.Join(fullPkgPath, packageJsonFileName))
	if err != nil {
		return "", errors.New("Couldn't get " + nodeModulesFolder + " path: " + err.Error())
	}
	//This isn't a relativized path because the UserCacheDir is not  nearby
	//this dir, unlike the game folders. So leave it as an absolute path.
	if err := os.Symlink(absRemoteNodePath, filepath.Join(staticDir, nodeModulesFolder)); err != nil {
		return "", errors.New("Couldn't symlink in " + nodeModulesFolder + ": " + err.Error())
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

func absoluteStaticServerPath() (string, error) {

	pth, err := path.AbsoluteGoPkgPath(mainPackage)

	if err != nil {
		return "", errors.New("Couldn't load main boardgame package location: " + err.Error())
	}

	return filepath.Join(pth, staticServerPath), nil

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

//buildCachePath returns where we store our build cache (or where we WOULD if
//it existed).
func buildCachePath() (string, error) {
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", errors.New("Couldn't get usercachedir: " + err.Error())
	}

	return filepath.Join(userCacheDir, nodeModulesCacheDir), nil
}

//updateNodeModules returns an absolute path to where on disk the node_modules
//folder for the static resources is. Takes an absolute path to the
//package.json to use. If it doesn't exist it will create it and update. It
//will call `npm up` on it even if it already exists to ensure it is up to
//date. The node_modules will be stored in a user cache dir.
func updateNodeModules(absPackageJsonPath string) (string, error) {

	_, err := exec.LookPath("npm")

	if err != nil {
		return "", errors.New("npm didn't appear to be installed. You need to install npm.")
	}

	cacheDir, err := buildCachePath()
	if err != nil {
		return "", errors.New("Couldn't get build cache path: " + err.Error())
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Println("Creating " + cacheDir + " You can remove it with boardgame-util clean cache")
		if err := os.Mkdir(cacheDir, 0700); err != nil {
			return "", errors.New("Couldn't create cache dir: " + err.Error())
		}
	}

	if _, err := os.Stat(absPackageJsonPath); os.IsNotExist(err) {
		return "", errors.New("The path to package.json didn't denote a real file")
	}

	//Copy over package.json
	if err := copyFile(absPackageJsonPath, filepath.Join(cacheDir, packageJsonFileName)); err != nil {
		return "", errors.New("Couldn't copy over package.json: " + err.Error())
	}

	nodeCacheDir := filepath.Join(cacheDir, nodeModulesFolder)

	if _, err := os.Stat(nodeCacheDir); os.IsNotExist(err) {
		fmt.Println("Downloading initial npm modules. This might take awhile, but future builds can skip it...")
	} else if err == nil {
		fmt.Println("node_modules already exists, doing quick `npm up` to make sure it's up to date...")
	}

	//call `npm up`, warning if it fails
	cmd := exec.Command("npm", "up")
	cmd.Dir = cacheDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	npmUpErrored := false

	if err := cmd.Run(); err != nil {
		//Don't quit because it's not NECESSARILY an error, if node_modules
		//already existed. For example, if they're on an airplane without wifi
		//and have node_modules already (even if out of date) it's OK.
		fmt.Println("WARNING: npm up failed: " + err.Error())
		npmUpErrored = true
	}

	if _, err := os.Stat(nodeCacheDir); os.IsNotExist(err) {
		//As long as node_modules exists, even if `npm up` failed (perhaps
		//because the user is not on wifi), then it's fine.
		return "", errors.New("node_modules cache could not be created, aborting build")
	} else if err != nil {
		return "", errors.New("Unexpected error: " + err.Error())
	} else if npmUpErrored {
		fmt.Println("An older version of " + nodeModulesFolder + " still existed, so proceeding...")
	}

	return nodeCacheDir, nil

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
