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

	var err error

	switch mode {
	case "up":
		err = up()
	case "down":
		err = down()
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

func up() error {
	return errors.New("Not yet implemented")
}

func down() error {
	return errors.New("Not yet implemented")
}

func clean() error {
	return errors.New("Not yet implemented")
}
