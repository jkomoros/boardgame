package static

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"os"
	"os/exec"
	"path/filepath"
)

var filesToExclude map[string]bool = map[string]bool{
	".gitignore":      true,
	"README.md":       true,
	polymerConfig:     true,
	nodeModulesFolder: true,
	//Don't copy over because we'll generate our own; if we copy over and
	//generate our own we'll overwrite original.
	clientConfigJsFileName: true,
	".DS_Store":            true,
}

//Server runs a static server. directory is the folder that the `static`
//folder is contained within. If no error is returned, runs until the program
//exits. Under the cover uses `polymer serve` because imports use bare module
//specifiers that must be rewritten.
func Server(directory string, port string) error {

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

/*

Build creates a folder of static resources for a server in the given
directory. It is the primary entrypoint for this package. It has no logic of
its own but serves to call all of the build steps in the correct order.

Specificlaly, it calls: CopyStaticResources, passing copyFiles;
LinkNodeModules; CreateClientConfigJs, passing c; LinkGameClientFolders,
passing gameImports; CreatePolymerJson, passing false. If prodBuild is true,
also calls BuildPolymer.

See the package doc for more about the specific build steps and what they do.

*/
func Build(directory string, gameImports []string, c *config.Config, prodBuild bool, copyFiles bool) (assetRoot string, err error) {

	staticDir, err := staticBuildDir(directory)
	if err != nil {
		return "", err
	}

	fmt.Println("Copying base static resources")
	if err := CopyStaticResources(directory, copyFiles); err != nil {
		return "", errors.New("Couldn't copy static resources")
	}

	fmt.Println("Updating " + nodeModulesFolder + " and linking in")
	if err := LinkNodeModules(directory); err != nil {
		return "", errors.New("Couldn't link " + nodeModulesFolder + ": " + err.Error())
	}

	fmt.Println("Creating " + clientConfigJsFileName)
	if err := CreateClientConfigJs(directory, c); err != nil {
		return "", errors.New("Couldn't create " + clientConfigJsFileName + ": " + err.Error())
	}

	fmt.Println("Creating " + gameSrcSubFolder)
	if err := LinkGameClientFolders(directory, gameImports); err != nil {
		return "", errors.New("Couldn't create " + gameSrcSubFolder + ": " + err.Error())
	}

	fmt.Println("Creating " + polymerConfig)
	if err := CreatePolymerJson(directory, false); err != nil {
		return "", errors.New("Couldn't create " + polymerConfig + ": " + err.Error())
	}

	if prodBuild {
		fmt.Println("Building bundled resources with `polymer build`")
		if err := BuildPolymer(directory); err != nil {
			return "", errors.New("Couldn't build bundled resources: " + err.Error())
		}
	}

	return staticDir, nil

}

//Clean removes all of the things created in the static subfolder within
//directory.
func Clean(directory string) error {
	return os.RemoveAll(filepath.Join(directory, staticSubFolder))
}
