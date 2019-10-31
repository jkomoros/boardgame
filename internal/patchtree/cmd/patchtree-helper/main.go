/*

patchtree-helper is a simple command that wraps patchtree.ExpandTree and ContractTree.

It is useful to modify base.json in a patchtree. The workflow is: sitting
in the directory with base.json, run `patchtree-helper expand`. Then
modify base.json. Then run `patchtree-helper contract` to generate new
`modification.patch`. Then run `patchtree-helper clean` to remove the
temporary `node.expanded.json`.

*/
package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/internal/patchtree"
)

const validModesMessage = "Valid modes are 'expand', 'contract', 'clean'."

func main() {
	if len(os.Args) < 2 {
		log.Println("Need to provide a mode. " + validModesMessage)
		os.Exit(1)
	}

	mode := strings.ToLower(os.Args[1])

	dir, err := os.Getwd()

	if err != nil {
		log.Println("couldn't get working directory: " + err.Error())
		os.Exit(1)
	}

	log.Println(dir)

	var affectedFiles int

	switch mode {
	case "expand":
		affectedFiles, err = patchtree.ExpandTree(dir)
	case "contract":
		affectedFiles, err = patchtree.ContractTree(dir)
	case "clean":
		affectedFiles, err = patchtree.CleanTree(dir)
	default:
		err = errors.New("Invalid mode provided. " + validModesMessage)
	}

	if err != nil {
		log.Println("Error: " + err.Error())
		os.Exit(1)
	}

	log.Println("Affected " + strconv.Itoa(affectedFiles) + " files.")

	return

}
