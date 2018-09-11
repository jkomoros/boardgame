package static

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const polymerConfig = "polymer.json"
const polymerFragmentsKey = "fragments"

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

//CreatePolymerJson outputs a polymer.json in the given build folder that
//includes fragments for all of the valid game directories that exist in the
//build folder's game-src directory. If missingFragmentsError is true, then if
//there are missing fragments we will return an error. Otherwise, will print a
//message for missing fragments but not error.
func CreatePolymerJson(dir string, missingFragmentsErrors bool) error {

	staticDir, err := staticBuildDir(dir)
	if err != nil {
		return errors.New("Couldn't get static di: " + err.Error())
	}

	path := filepath.Join(staticDir, polymerConfig)

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

//BuildPolymer calls `polymer build` to build the bundled version within a
//given build directory. Will error if the build directory doesn't include
//polymer.json (which CreatePolymerJson will have created).
func BuildPolymer(dir string) error {

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
