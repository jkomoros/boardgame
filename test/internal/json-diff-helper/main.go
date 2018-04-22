/*

	json-diff-helper is a simple little utility helper that makes it easier to
	modify expected json outputs for states.

	Run it in boardgame/test.

	Three modes:

	up : For each .patch file foo.patch, creates a jd patch against basic_state.json, and saves it as foo.temp.json
	down : For each foo.temp.json file, rediff against the new basic_state.json, and overwrite foo.patch with the new patch.
	clean : Remove each foo.temp.json file

	The workflow is run `json-diff-helper up`, then modify basic_state.json,
	then run `json-diff-helper down`, then verify it looks correct, then run
	`json-diff-helper clean`.

*/
package main

import (
	"errors"
	jd "github.com/jkomoros/jd/lib"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const validModesMessage = "Valid modes are 'up', 'down', and 'clean'."

func main() {
	if len(os.Args) < 2 {
		log.Println("Need to provide a mode. " + validModesMessage)
		os.Exit(1)
	}

	mode := strings.ToLower(os.Args[1])

	baseFile := "basic_state.json"

	if len(os.Args) >= 3 {
		baseFile = os.Args[2]
	} else {
		log.Println("No base json file provided. Defaulting to " + baseFile)
	}

	if _, err := os.Stat(baseFile); os.IsNotExist(err) {
		log.Println("Provided base json file (" + baseFile + ") does not exist.")
		os.Exit(1)
	}

	var err error

	switch mode {
	case "up":
		err = up(baseFile)
	case "down":
		err = down(baseFile)
	case "clean":
		err = clean()
	default:
		err = errors.New("Invalid mode provided. " + validModesMessage)
	}

	if err != nil {
		log.Println("Error: " + err.Error())
		os.Exit(1)
	}

	return

}

func up(baseFile string) error {

	files, err := ioutil.ReadDir(".")

	if err != nil {
		return errors.New("Couldn't read dir: " + err.Error())
	}

	baseJson, err := jd.ReadJsonFile(baseFile)

	if err != nil {
		return errors.New("Couldn't parse base json file: " + err.Error())
	}

	//We'll keep track of results before saving, to make sure there are no
	//errors before we save.
	result := make(map[string]jd.JsonNode)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".patch") {
			continue
		}

		diff, err := jd.ReadDiffFile(file.Name())

		if err != nil {
			return errors.New("Couldn't read diff file " + file.Name() + ": " + err.Error())
		}

		patchedJson, err := baseJson.Patch(diff)

		if err != nil {
			return errors.New("Couldn't patch diff file " + file.Name() + ": " + err.Error())
		}

		result[strings.Replace(file.Name(), ".patch", ".temp.json", -1)] = patchedJson
	}

	for fileName, node := range result {
		if err := ioutil.WriteFile(fileName, []byte(node.Json()), 0644); err != nil {
			return errors.New("Couldn't write file: " + fileName + ": " + err.Error())
		}
	}

	return nil

}

func down(baseFile string) error {
	return errors.New("Not yet implemented")
}

func clean() error {
	return errors.New("Not yet implemented")
}
