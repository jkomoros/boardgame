/*

	gamepkg is a package that helps locate, validate, and modify game package
	imports.

*/
package gamepkg

import (
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"io/ioutil"
	"os"
	"path/filepath"
)

type GamePkg struct {
	//Every contstructo sets absolutePath to something that at least exists on
	//disk.
	absolutePath string
	importPath   string
}

//NewFromPath takes path (either relative or absolute path) and returns a new
//GamePkg. Will error if the given path does not appear to denote a valid game package for any reason.
func NewFromPath(path string) (*GamePkg, error) {

	if !filepath.IsAbs(path) {

		cwd, err := os.Getwd()

		if err != nil {
			return nil, errors.New("Couldn't get working directory: " + err.Error())
		}

		path = filepath.Join(cwd, path)
	}

	return newGamePkg(path, "")

}

//NewFromImport will return a new GamePkg pointing to that import.
func NewFromImport(importPath string) (*GamePkg, error) {

	absPath, err := path.AbsoluteGoPkgPath(importPath)

	if err != nil {
		return nil, errors.New("Absolute path couldn't be found: " + err.Error())
	}

	//If no error, then absPath must point to a valid thing

	return newGamePkg(absPath, importPath)

}

func newGamePkg(absPath, importPath string) (*GamePkg, error) {
	result := &GamePkg{
		absolutePath: absPath,
		importPath:   importPath,
	}

	if info, err := os.Stat(absPath); err != nil {
		return nil, errors.New("Path doesn't point to valid location on disk: " + err.Error())
	} else if !info.IsDir() {
		return nil, errors.New("Path points to an object but it's not a directory.")
	}

	if !result.goPkg() {
		return nil, errors.New(absPath + " denotes a folder with no go source files")
	}

	return result, nil
}

//AbsolutePath returns the absolute path where the package in question resides
//on disk. All constructors will have errored if AbsolutePath doesn't at the
//very least point to a valid location on disk.
func (g *GamePkg) AbsolutePath() string {
	return g.absolutePath
}

//goPkg validates that the absolutePath denotes a package with at least one go
//file. If there's an error will default to false.
func (g *GamePkg) goPkg() bool {

	infos, _ := ioutil.ReadDir(g.AbsolutePath())

	for _, info := range infos {
		if filepath.Ext(info.Name()) == ".go" {
			return true
		}
	}

	return false

}
