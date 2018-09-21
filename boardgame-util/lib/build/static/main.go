/*

	static is a library that helps automate creating the static directory of
	files for the webapp to function. It, along with lib/build/api, is one
	half of the build process to produce a functioning webapp given a
	config.json

	All of the steps it does could be done by hand, but they are finicky and
	error prone, so this library helps take the guess work out of them.

	Build() is the primary entrypoint for this package, which composes all of
	the build steps, which are themselves exposed as public methods int his
	package, in the right order.

	All build methods in this package take a dir parameter. This is the
	directory to produce the build into. In particular, these methods will
	create a `static` subdirectory in dir and work inside of that. dir may be
	"", which will create the static sub-folder within the current directory.
	Tools like `boardgame-util serve` create a temporary directory and use
	that, so it's easy to clean up later.

	Various steps of the build symlink other files and folders into the
	created static build directory, for example when we symlink in the client
	directories of each game into game_src. These symlinks use relative paths,
	not absolute paths. This means that if all of the packages (including the
	current repo) are in the canonical location in $GOPATH (and modules are
	not enabled), then these relative symlinks are OK to check in to a source
	control repo because they should work reasonably on other systems.
	However, typically you don't check in the results of the static build into
	source control, and instead use `boardgame-util serve`, which creates a
	temporary directory for the build each time.

	The steps of the build process, at a high level, are as follows:

	First, create the `static` sub directory, if it doesn't already exist. All
	following steps create files and directories within that static subfolder.

	Next, it copies over all of the static resources (no directories) from
	`github.com/jkomoros/boardgame/server/static`, skipping a handful of files
	that will be generated later. These files are symlinked by default, but
	can also be copied. This step is encapsulated by CopyStaticResources.

	Next, it creates a node_modules folder that contains up to date
	dependencies given the contents of
	`github.com/jkomoros/boardgame/server/static/package.json`. Checking out
	this whole directory is expensive, so this package creates a node_modules
	in a central cache, re-upping it each time this command is run (unless
	skipUpdate is true), and then symlinks it into the static directory. This
	step is encapsulated by LinkNodeModules.

	Next, it generates a `client_config.js`, which encodes the global
	configuration for the client webapp. It calls config.Client(false) and
	saves the result to static/client-config.js, which index.html will look
	for when booting up. This step is encapsulated by CreateClientConfigJs.

	Next, it copies in the client folders (containing boardgame-render-game-
	GAMENAME.js, and optionally boardgame-render-player-info-GAMENAME.js) into
	static/game-src. It does this by locating the on-disk location of each
	gameImport given by gameImports (typically this is configMode.Games), then
	symlinking its client folder into `static/game-src/GAMENAME`. In a modules
	context, game packages that do not yet exist on disk will be downloaded
	automatically; if you are not using modules and you have not yet `go get`
	the given game imports or a containg package, it will error. This step is
	encapsulated by LinkGameClientFolders.

	Next, it generates a `static/polymer.json`, which contains fragments
	entries for each dynamic import--specifically, the `game-src/GAMENAME
	/boardgame-render-game-GAMENAME.js` and `game-src/GAMENAME/boardgame-
	render-player-info-GAMENAME.js`, if it exists. It identifes the fragments
	to include by walking through all of the game directories in `game-src`,
	meaning it relies on the output of the previous step. It then saves this
	generated file to `static/polymer.json`. This step is encapsulated by
	CreatePolymerJson.

	The static build is now mostly complete. Optionally, BuildPolymer can be
	called to run `polymer build` on the generated static dir. This step is
	encapsulated by BuildPolymer.

	Typically direct users of this package use Build(), which automatically
	runs these steps in the proper order.

	Clean() removes the static build contents from the given build directory
	(specifically, it removes the `static` subdirectory and all of its
	contents). LinkNodeModules also might create (or update) a shared cache
	directory of node_modules on the system, and CleanCache() removes that
	cache.

	Server() is a simple development server that makes the static resources
	available at `localhost:PORT`. Polymer requires that imports use bare
	module specifiers, which means that a naive local serve is not sufficient
	because the import URLs must be lightly rewritten, so this Server() is
	necessary. Under the covers it uses `polymer serve`.

	Typically you don't use this package directly, but use `boardgame-util
	build static` or `boardgame-util serve`.

*/
package static

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
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
LinkNodeModules, passing skipNodeUpdate; CreateClientConfigJs, passing c;
LinkGameClientFolders, passing gameImports; CreatePolymerJson, passing false.
If prodBuild is true, also calls BuildPolymer.

See the package doc for more about the specific build steps and what they do.

*/
func Build(directory string, pkgs []*gamepkg.Pkg, c *config.Config, prodBuild bool, copyFiles bool, skipNodeUpdate bool) (assetRoot string, err error) {

	staticDir, err := staticBuildDir(directory)
	if err != nil {
		return "", err
	}

	fmt.Println("Copying base static resources")
	if err := CopyStaticResources(directory, copyFiles); err != nil {
		return "", errors.New("Couldn't copy static resources")
	}

	fmt.Println("Updating " + nodeModulesFolder + " and linking in")
	if err := LinkNodeModules(directory, skipNodeUpdate); err != nil {
		return "", errors.New("Couldn't link " + nodeModulesFolder + ": " + err.Error())
	}

	fmt.Println("Creating " + clientConfigJsFileName)
	if err := CreateClientConfigJs(directory, c); err != nil {
		return "", errors.New("Couldn't create " + clientConfigJsFileName + ": " + err.Error())
	}

	fmt.Println("Creating " + gameSrcSubFolder)
	if err := LinkGameClientFolders(directory, pkgs); err != nil {
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
