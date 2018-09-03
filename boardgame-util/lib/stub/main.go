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

//Options is the default options struct. Name is the only required field; the
//zero-value of every other field is default.
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
	SuppressTest  bool
	SuppressPhase bool
	//If true, won't add a CurrentPlayer to gameState
	SuppressCurrentPlayer bool
	//If true, won't render moves_normal.go, moves_setup.go, or moves.go if suppressphase is true
	SuppressMovesStubs                  bool
	SuppressComponentsStubs             bool
	SuppressClientRenderGame            bool
	SuppressClientRenderPlayerInfo      bool
	EnableExampleDeck                   bool
	EnableExampleDynamicComponentValues bool
	EnableExampleEndState               bool
	EnableExampleComputedProperties     bool
	EnableExampleConstants              bool
	EnableExampleConfigs                bool
	EnableExampleClient                 bool
	EnableExampleMoves                  bool
}

//FileContents is the generated contents of the files to later write to the
//filesystem.
type FileContents map[string][]byte

//SuppressClient sets the options about clients to suppressed.
func (o *Options) SuppressClient() {
	o.SuppressClientRenderGame = true
	o.SuppressClientRenderPlayerInfo = true
}

//SuppressExtras sets all of the non-client extras that are on by default to
//off.
func (o *Options) SuppressExtras() {
	o.SuppressTest = true
	o.SuppressPhase = true
	o.SuppressCurrentPlayer = true
	o.SuppressMovesStubs = true
	o.SuppressComponentsStubs = true
}

//EnableTutorials enables all of the off-by-default tutorial options.
func (o *Options) EnableTutorials() {
	o.EnableExampleDeck = true
	o.EnableExampleDynamicComponentValues = true
	o.EnableExampleEndState = true
	o.EnableExampleComputedProperties = true
	o.EnableExampleConstants = true
	o.EnableExampleConfigs = true
	o.EnableExampleClient = true
	o.EnableExampleMoves = true
}

//Validate verifies that Options is in a legal state. Makes sure Name exists
//and ensures it's lowerCase. Repeated calls are OK.
func (o *Options) Validate() error {

	o.Name = strings.TrimSpace(o.Name)
	o.Name = strings.ToLower(o.Name)

	if o.Name == "" {
		return errors.New("No name provided")
	}

	if o.MinNumPlayers != 0 && o.MaxNumPlayers != 0 {
		if o.MaxNumPlayers < o.MinNumPlayers {
			return errors.New("Max num players less than min")
		}

		if o.DefaultNumPlayers != 0 {
			if o.DefaultNumPlayers < o.MinNumPlayers || o.DefaultNumPlayers > o.MaxNumPlayers {
				return errors.New("Default num players not within min/max range")
			}
		}
	}

	//We don't verify that the name is fully legal according to the boardgame
	//framework, because that test will fail given the test generated in
	//main_test.go.

	//Disallow illegal combinations
	if o.EnableExampleDynamicComponentValues {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleEndState {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleComputedProperties {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleConstants {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleClient {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleMoves {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleMoves {
		o.SuppressPhase = false
	}

	if o.EnableExampleClient {
		o.SuppressClientRenderGame = false
		o.SuppressClientRenderPlayerInfo = false
	}

	if o.EnableExampleDeck {
		o.SuppressComponentsStubs = false
	}

	return nil
}

//Generate generates FileContents for the given set of options. A convenience
//wrapper around DefaultTemplateSet, templates.Generate(), and files.Format().
func Generate(opt *Options) (FileContents, error) {

	if err := opt.Validate(); err != nil {
		return nil, errors.New("Options didn't validate: " + err.Error())
	}

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
