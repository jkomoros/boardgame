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
	"os"
	"path/filepath"
)

const BASE_JSON = "base.json"
const PATCH = "modification.patch"

//JSON returns the patched json blob impplied by that directory structure or
//an error if something doesn't work. See the package doc for more.
func JSON(path string) ([]byte, error) {

	result, err := processDirectory(path)

	if err != nil {
		return nil, err
	}

	return []byte(result.Json()), nil

}

func processDirectory(path string) (jd.JsonNode, error) {

	//If no more path pieces error
	if path == "" || path == "/" || path == "./" {
		return nil, errors.New("Didn't find a base.json anywhere in the given directory structure")
	}

	//TODO: check if the directory exists...

	baseJsonPath := filepath.Clean(path + "/" + BASE_JSON)

	if _, err := os.Stat(baseJsonPath); os.IsExist(err) {
		//Found the directory with base.json!
		return jd.ReadJsonFile(baseJsonPath)
	}

	modificationPatchPath := filepath.Clean(path + "/" + PATCH)

	if _, err := os.Stat(modificationPatchPath); os.IsExist(err) {

		//Recurse, with the sub-directory.
		baseJson, err := processDirectory(filepath.Dir(path))

		if err != nil {
			return nil, err
		}

		diff, err := jd.ReadDiffFile(modificationPatchPath)

		if err != nil {
			return nil, errors.New("Error reading diff file at " + modificationPatchPath + ": " + err.Error())
		}

		return baseJson.Patch(diff)
	}

	//Path had neither base.json or modification.patch, which is an error
	return nil, errors.New("In " + path + "didn't have either " + BASE_JSON + " or " + PATCH)

}
