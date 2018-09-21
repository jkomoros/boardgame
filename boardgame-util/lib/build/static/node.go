package static

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const nodeModulesFolder = "node_modules"
const packageJsonFileName = "package.json"

//The name of the direcotry within os.UserCacheDir() that node_modules should
//be created within.
const nodeModulesCacheDir = "com.github.jkomoros.boardgame"

//LinkNodeModules symlinks a node_modules folder into the build directory
//given by dir, that is fully up to date based on the resources required.
//node_modules is cached in a known cache on the system, and only topped up as
//necessary, so only the first call to this on a given system should be
//particularly expensive (or after CleanCache()) has been called. Returns an
//error if node_modules can't be updated or if it can't be linked in. If
//skipUpdate is true, then if node_modules exists we won't try to update.
//Useful if you're in an offline context.
func LinkNodeModules(dir string, skipUpdate bool) error {

	staticDir, err := staticBuildDir(dir)

	if err != nil {
		return errors.New("Couldn't get static dir: " + err.Error())
	}

	fullPkgPath, err := absoluteStaticServerPath()
	if err != nil {
		return errors.New("Couldn't get full package path: " + err.Error())
	}

	//Ensure node_modules exists adn link to it
	absRemoteNodePath, err := updateNodeModules(filepath.Join(fullPkgPath, packageJsonFileName), skipUpdate)
	if err != nil {
		return errors.New("Couldn't get " + nodeModulesFolder + " path: " + err.Error())
	}
	//This isn't a relativized path because the UserCacheDir is not  nearby
	//this dir, unlike the game folders. So leave it as an absolute path.
	if err := os.Symlink(absRemoteNodePath, filepath.Join(staticDir, nodeModulesFolder)); err != nil {
		return errors.New("Couldn't symlink in " + nodeModulesFolder + ": " + err.Error())
	}

	return nil
}

//updateNodeModules returns an absolute path to where on disk the node_modules
//folder for the static resources is. Takes an absolute path to the
//package.json to use. If it doesn't exist it will create it and update. It
//will call `npm up` on it even if it already exists to ensure it is up to
//date. The node_modules will be stored in a user cache dir. If skipUpdate is
//true, then if the folder already exists it will skip it.
func updateNodeModules(absPackageJsonPath string, skipUpdate bool) (string, error) {

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

	nodeCacheExists := true

	if _, err := os.Stat(nodeCacheDir); os.IsNotExist(err) {
		fmt.Println("Downloading initial npm modules. This might take awhile, but future builds can skip it...")
		nodeCacheExists = false
	} else if err == nil {
		fmt.Println("node_modules already exists, doing quick `npm up` to make sure it's up to date...")
	}

	if skipUpdate {
		if !nodeCacheExists {
			return "", errors.New("Node cache didn't exist, but we were told not to update node. Aborting build.")
		}
		fmt.Println("node_modules existed, but we were told not to update so skipping `npm up`")
		return nodeCacheDir, nil
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
