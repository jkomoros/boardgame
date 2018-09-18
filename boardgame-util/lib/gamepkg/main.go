/*

	gamepkg is a package that helps locate, validate, and modify game package
	imports.

*/
package gamepkg

import (
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
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
//GamePkg. It only returns an error if path doesn't denote a real object on
//disk; it doesn't do any other checking yet.
func NewFromPath(path string) (*GamePkg, error) {

	if !filepath.IsAbs(path) {

		cwd, err := os.Getwd()

		if err != nil {
			return nil, errors.New("Couldn't get working directory: " + err.Error())
		}

		path = filepath.Join(cwd, path)
	}

	if info, err := os.Stat(path); err != nil {
		return nil, errors.New("Path doesn't point to valid location on disk: " + err.Error())
	} else if !info.IsDir() {
		return nil, errors.New("Path points to an object but it's not a directory.")
	}

	result := &GamePkg{
		absolutePath: path,
	}

	return result, nil

}

//NewFromImport will return a new GamePkg pointing to that import.
func NewFromImport(importPath string) (*GamePkg, error) {

	absPath, err := path.AbsoluteGoPkgPath(importPath)

	if err != nil {
		return nil, errors.New("Absolute path couldn't be found: " + err.Error())
	}

	//If no error, then absPath must point to a valid thing

	result := &GamePkg{
		absolutePath: absPath,
		importPath:   importPath,
	}

	return result, nil

}

//AbsolutePath returns the absolute path where the package in question resides
//on disk. All constructors will have errored if AbsolutePath doesn't at the
//very least point to a valid location on disk.
func (g *GamePkg) AbsolutePath() string {
	return g.absolutePath
}
