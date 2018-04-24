/*

	patchtree is a simple library that knows how to interpret a folder
	structure of jd diffs and apply them on top of a base.

		test/
		  base.json
		  after_move/
		    modification.patch
		  sanitization/
		    modification.patch
		    hidden/
		      modification.patch
		    nonempty/
		      modification.patch

	Given a path relative ot the current binary, it walks backwards up the
	folder, ensuring that a modification.patch exists in each directory until
	it finds a base.json. Then it applies forward all of the
	modification.patches to give you the final composed json blob result.

*/
package patchtree

import (
	"errors"
	jd "github.com/jkomoros/jd/lib"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const BASE_JSON = "base.json"
const PATCH = "modification.patch"

//JSON returns the patched json blob impplied by that directory structure or
//an error if something doesn't work. See the package doc for more.
func JSON(path string) ([]byte, error) {

	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	result, err := processDirectory(path)

	if err != nil {
		return nil, err
	}

	return []byte(result.Json()), nil

}

//MustJSON is the same as JSON, but if it would have returned an error, panics istead.
func MustJSON(path string) []byte {
	result, err := JSON(path)
	if err != nil {
		panic(err)
	}
	return result
}

//ExpandTree expands all of the nodes in the patchtree, applying the chains of
//modification and created an node.expanded.json in each node. Used in a
//workflow to modify base.json: run this commeand, then modify base.json, then
//run ContractTree.
func ExpandTree(rootPath string) error {

	baseJsonPath := filepath.Clean(rootPath + "/" + BASE_JSON)

	if _, err := os.Stat(baseJsonPath); os.IsNotExist(err) {
		return errors.New("Base json file did not exist: " + err.Error())
	}

	node, err := jd.ReadJsonFile(baseJsonPath)

	if err != nil {
		return errors.New("Couldn't parse base json file: " + err.Error())
	}

	files, err := ioutil.ReadDir(rootPath)

	if err != nil {
		return errors.New("Couldn't read base directory: " + err.Error())
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		if err := expandTreeProcessDirectory(filepath.Clean(rootPath+"/"+file.Name()), node); err != nil {
			return errors.New("Couldn't process file: " + err.Error())
		}
	}

	return nil

}

func expandTreeProcessDirectory(directory string, node jd.JsonNode) error {
	return errors.New("Subdirectories not yet implemented")
}

//ContractTree goes through each node in the parse tree and where it finds a
//node.expanded,json, re-derives and overwrites the "modification.patch". Used
//as part of a workflow to modify base.json: run ExpandTree, modify base.json,
//then ContractTree.
func ContractTree(rootPath string) error {
	return errors.New("Not yet implemented")
}

func processDirectory(path string) (jd.JsonNode, error) {

	//If no more path pieces error
	if path == "" || path == "/" || path == "./" {
		return nil, errors.New("Didn't find a base.json anywhere in the given directory structure")
	}

	//TODO: check if the directory exists...

	baseJsonPath := filepath.Clean(path + "/" + BASE_JSON)

	if _, err := os.Stat(baseJsonPath); err == nil {
		//Found the directory with base.json!
		node, err := jd.ReadJsonFile(baseJsonPath)
		if err != nil {
			return nil, errors.New(path + " had error reading base.json: " + err.Error())
		}
		return node, nil
	}

	modificationPatchPath := filepath.Clean(path + "/" + PATCH)

	if _, err := os.Stat(modificationPatchPath); err == nil {

		//Recurse, with the sub-directory.
		baseJson, err := processDirectory(filepath.Dir(path))

		if err != nil {
			return nil, err
		}

		diff, err := jd.ReadDiffFile(modificationPatchPath)

		if err != nil {
			return nil, errors.New("Error reading diff file at " + modificationPatchPath + ": " + err.Error())
		}

		composed, err := baseJson.Patch(diff)

		if err != nil {
			return nil, errors.New(path + " had error diffing " + err.Error())
		}

		return composed, nil
	}

	//Path had neither base.json or modification.patch, which is an error
	return nil, errors.New("In " + path + " didn't have either " + BASE_JSON + " or " + PATCH)

}
