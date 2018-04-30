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

	patchtree-helper is a simple cli utility that wraps the functionality in
	this package.

*/
package patchtree

import (
	"encoding/json"
	"errors"
	jd "github.com/jkomoros/jd/lib"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//BATCH_JSON_NAME is the name of the base json file at the root of the
//directory tree.
const BASE_JSON_NAME = "base.json"

//PATCH_NAME is the name of the patch in each sub-folder that modifies the
//json in the tree above it.
const PATCH_NAME = "modification.patch"

//EXPANDED_JSON_NAME is the name of the file that represents the entire
//expanded json blob at a given part of the tree, created by `expand`.
const EXPANDED_JSON_NAME = "node.expanded.json"

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

type directoryFunc func(string, jd.JsonNode) (int, error)

func startDirectoryAndWalk(rootPath string, subFunc directoryFunc) (int, error) {
	baseJsonPath := filepath.Clean(rootPath + "/" + BASE_JSON_NAME)

	if _, err := os.Stat(baseJsonPath); os.IsNotExist(err) {
		return 0, errors.New("Base json file did not exist: " + err.Error())
	}

	node, err := jd.ReadJsonFile(baseJsonPath)

	if err != nil {
		return 0, errors.New("Couldn't parse base json file: " + err.Error())
	}

	return walkDirectory(rootPath, node, subFunc)
}

func walkDirectory(directory string, expandedNode jd.JsonNode, subFunc directoryFunc) (int, error) {
	files, err := ioutil.ReadDir(directory)

	if err != nil {
		return 0, errors.New("Couldn't read directory: " + err.Error())
	}

	numAffectedFiles := 0

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		subDirectory := filepath.Clean(directory + "/" + file.Name())

		subAffectedFiles, err := subFunc(subDirectory, expandedNode)
		if err != nil {
			return numAffectedFiles, err
		}
		numAffectedFiles += subAffectedFiles
	}

	return numAffectedFiles, nil
}

//ExpandTree expands all of the nodes in the patchtree, applying the chains of
//modification and created an node.expanded.json in each node. Used in a
//workflow to modify base.json: run this commeand, then modify base.json, then
//run ContractTree.
func ExpandTree(rootPath string) (affectedFiles int, err error) {
	return startDirectoryAndWalk(rootPath, expandTreeProcessDirectory)
}

func expandTreeProcessDirectory(directory string, node jd.JsonNode) (int, error) {

	diffFileName := filepath.Clean(directory + "/" + PATCH_NAME)

	if _, err := os.Stat(diffFileName); os.IsNotExist(err) {
		//TODO: it's weird to print to log when this condition is hit.
		log.Println(diffFileName + " did not exist; skipping that directory and all beneath it.")
		return 0, nil
	}

	diff, err := jd.ReadDiffFile(diffFileName)

	if err != nil {
		return 0, errors.New(diffFileName + " could not be loaded as patch file: " + err.Error())
	}

	expandedNode, err := node.Patch(diff)

	if err != nil {
		return 0, errors.New(diffFileName + " could not be applied: " + err.Error())
	}

	expandedNodeFileName := filepath.Clean(directory + "/" + EXPANDED_JSON_NAME)

	data := expandedNode.Json()

	indentedJson, err := indentJson(data)

	if err != nil {
		return 0, errors.New("Couldn't indent json: " + err.Error())
	}

	if err := ioutil.WriteFile(expandedNodeFileName, []byte(indentedJson), 0644); err != nil {
		return 0, errors.New("Couldn't write " + expandedNodeFileName + ": " + err.Error())
	}

	numAffectedFiles, err := walkDirectory(directory, expandedNode, expandTreeProcessDirectory)

	if err != nil {
		return 0, err
	}

	return numAffectedFiles + 1, nil

}

func indentJson(data string) (indented string, err error) {
	var obj map[string]interface{}

	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return "", errors.New("Couldn't unpack generated json: " + err.Error())
	}

	result, err := json.MarshalIndent(obj, "", "\t")

	if err != nil {
		return "", errors.New("Couldn't repack generated json: " + err.Error())
	}

	return string(result), nil

}

//ContractTree goes through each node in the parse tree and where it finds a
//node.expanded,json, re-derives and overwrites the "modification.patch". Used
//as part of a workflow to modify base.json: run ExpandTree, modify base.json,
//then ContractTree.
func ContractTree(rootPath string) (numAffectedFiles int, err error) {
	return startDirectoryAndWalk(rootPath, contractTreeProcessDirectory)
}

func contractTreeProcessDirectory(directory string, node jd.JsonNode) (int, error) {

	nodeFileName := filepath.Clean(directory + "/" + EXPANDED_JSON_NAME)

	expandedNode, err := jd.ReadJsonFile(nodeFileName)

	if err != nil {
		return 0, errors.New(nodeFileName + " could not be loaded as json file: " + err.Error())
	}

	patch := node.Diff(expandedNode)

	data := patch.Render()

	diffFileName := filepath.Clean(directory + "/" + PATCH_NAME)

	if err := ioutil.WriteFile(diffFileName, []byte(data), 0644); err != nil {
		return 0, errors.New("Couldn't write " + diffFileName + ": " + err.Error())
	}

	numAffectedFiles, err := walkDirectory(directory, expandedNode, contractTreeProcessDirectory)

	if err != nil {
		return 0, err
	}

	return numAffectedFiles + 1, nil

}

//CleanTree goes through each node, and if modification.json conceptually
//matches node.expanded.json, then removes node.expanded.json.
func CleanTree(rootPath string) (numAffectedFiles int, err error) {
	return startDirectoryAndWalk(rootPath, cleanTreeProcessDirectory)
}

func cleanTreeProcessDirectory(directory string, node jd.JsonNode) (int, error) {

	nodeFileName := filepath.Clean(directory + "/" + EXPANDED_JSON_NAME)

	expandedNode, err := jd.ReadJsonFile(nodeFileName)

	if err != nil {
		return 0, errors.New(nodeFileName + " could not be loaded as json file: " + err.Error())
	}

	patchFileName := filepath.Clean(directory + "/" + PATCH_NAME)

	patch, err := jd.ReadDiffFile(patchFileName)

	if err != nil {
		return 0, errors.New(patchFileName + " could not be loaded as diff file: " + err.Error())
	}

	expandedNodeWithPatch, err := node.Patch(patch)

	if err != nil {
		return 0, errors.New(directory + " patch could not be applied: " + err.Error())
	}

	if !expandedNodeWithPatch.Equals(expandedNode) {
		return 0, errors.New(directory + " patch file did not match expanded json")
	}

	if err := os.Remove(nodeFileName); err != nil {
		return 0, errors.New("Couldn't delete node file: " + err.Error())
	}

	numAffectedFiles, err := walkDirectory(directory, expandedNode, cleanTreeProcessDirectory)

	if err != nil {
		return 0, err
	}

	return numAffectedFiles + 1, nil

}

func processDirectory(path string) (jd.JsonNode, error) {

	//If no more path pieces error
	if path == "" || path == "/" || path == "./" {
		return nil, errors.New("Didn't find a base.json anywhere in the given directory structure")
	}

	//TODO: check if the directory exists...

	baseJsonPath := filepath.Clean(path + "/" + BASE_JSON_NAME)

	if _, err := os.Stat(baseJsonPath); err == nil {
		//Found the directory with base.json!
		node, err := jd.ReadJsonFile(baseJsonPath)
		if err != nil {
			return nil, errors.New(path + " had error reading base.json: " + err.Error())
		}
		return node, nil
	}

	modificationPatchPath := filepath.Clean(path + "/" + PATCH_NAME)

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
	return nil, errors.New("In " + path + " didn't have either " + BASE_JSON_NAME + " or " + PATCH_NAME)

}
