/*

	stub is a library that helps generate stub code for new games

*/
package stub

import (
	"errors"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	//The name of the game
	Name string
	//DisplayName to output (skipped if "")
	DisplayName string
	//Description of game to output (skipped if "")
	Description       string
	MinNumPlayers     int
	MaxNumPlayers     int
	DefaultNumPlayers int
	//If true, won't save main_test.go
	SuppressTest bool
	//If true, won't add a CurrentPlayer to gameState
	SuppressCurrentPlayer bool
}

//FileContents is the generated contents of the files to later write to the
//filesystem.
type FileContents map[string][]byte

//Generate generates FileContents for the given set of options. A convenience
//wrapper around DefaultTemplateSet, templates.Generate(), and files.Format().
func Generate(opt *Options) (FileContents, error) {

	templates, err := DefaultTemplateSet(opt)

	if err != nil {
		return nil, errors.New("Default Template Set errored: " + err.Error())
	}

	if templates == nil {
		return nil, errors.New("No templates returned")
	}

	files, err := templates.Generate(opt)

	if err != nil {
		return nil, errors.New("Couldn't generate file contents: " + err.Error())
	}

	if err := files.Format(); err != nil {
		return nil, errors.New("Couldn't go fmt generated file contents: " + err.Error())
	}

	return files, nil
}

//Format go formats all of the code om FileContents whose path ends in ".go",
//erroring if the code isn't valid. If an error is returned, then the contents
//of FileContents will not have been modified.
func (f FileContents) Format() error {

	newContent := make(map[string][]byte)

	for filename, rawSource := range f {
		if strings.ToLower(filepath.Ext(filename)) != ".go" {
			continue
		}

		transformedSource, err := format.Source(rawSource)

		if err != nil {
			return errors.New("Couldn't format go code for " + filename + ": " + err.Error())
		}

		newContent[filename] = transformedSource
	}

	for name, content := range newContent {
		f[name] = content
	}

	return nil
}

//Save saves the given FileContents to the filesystem, creating any implied
//directories. Dir is the prefix to join with each path in FileContents; "" is
//fine. Will error if overwite is not true and any of the files to create
//already exist.
func (f FileContents) Save(dir string, overwrite bool) error {

	if !overwrite {
		for name := range f {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return errors.New(name + " already existed; save aborted")
			}
		}
	}

	for name := range f {
		path := filepath.Join(dir, name)
		dirsToCreate := filepath.Dir(path)
		if err := os.MkdirAll(dirsToCreate, os.ModePerm); err != nil {
			return errors.New("Couldn't create directories for " + path + ": " + err.Error())
		}
	}

	for name, contents := range f {
		path := filepath.Join(dir, name)

		if err := ioutil.WriteFile(path, contents, 0644); err != nil {
			return errors.New("Couldn't save " + path + ": " + err.Error())
		}
	}

	return nil
}

//InteractiveOptions renders an interactve prompt at out, in to generate an
//Options from the user. If in or out are nil, StdIn or StdOut will be used
//implicitly.
func InteractiveOptions(in, out *os.File) *Options {

	if in == nil {
		in = os.Stdin
	}

	if out == nil {
		out = os.Stdout
	}

	return nil
}
