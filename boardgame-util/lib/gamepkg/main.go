/*

	gamepkg is a package that helps locate, validate, and modify game package
	imports.

*/
package gamepkg

import (
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Pkg struct {
	//Every contstructo sets absolutePath to something that at least exists on
	//disk.
	absolutePath         string
	importPath           string
	calculatedIsGamePkg  bool
	memoizedIsGamePkg    bool
	memoizedIsGamePkgErr error
}

//New is a wrapper around NewFromImport and NewFromPath. First, it tries to
//interpret the input as an import. If that files, tries to interpret it as a
//path (rel or absolute), and if that fails, bails.
func New(importOrPath string) (*Pkg, error) {
	pkg, err := NewFromImport(importOrPath)
	if err == nil {
		return pkg, nil
	}
	return NewFromPath(importOrPath)
}

//NewFromPath takes path (either relative or absolute path) and returns a new
//Pkg. Will error if the given path does not appear to denote a valid game
//package for any reason.
func NewFromPath(path string) (*Pkg, error) {

	if !filepath.IsAbs(path) {

		cwd, err := os.Getwd()

		if err != nil {
			return nil, errors.New("Couldn't get working directory: " + err.Error())
		}

		path = filepath.Join(cwd, path)
	}

	return newPkg(path, "")

}

//NewFromImport will return a new Pkg pointing to that import. Will error
//if the given path does not appear to denote a valid game package for any
//reason.
func NewFromImport(importPath string) (*Pkg, error) {

	absPath, err := path.AbsoluteGoPkgPath(importPath)

	if err != nil {
		return nil, errors.New("Absolute path couldn't be found: " + err.Error())
	}

	//If no error, then absPath must point to a valid thing

	return newPkg(absPath, importPath)

}

func newPkg(absPath, importPath string) (*Pkg, error) {
	result := &Pkg{
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

	isGamePkg, err := result.isGamePkg()

	if !isGamePkg {
		return nil, errors.New(absPath + " was not a valid game package: " + err.Error())
	}

	return result, nil
}

//AbsolutePath returns the absolute path where the package in question resides
//on disk. All constructors will have errored if AbsolutePath doesn't at the
//very least point to a valid location on disk. For example, "/Users/YOURUSERNAME/Code/go/src/github.com/jkomoros/boardgame/examples/memory"
func (p *Pkg) AbsolutePath() string {
	return p.absolutePath
}

//ReadOnly returns true if the package appears to be in a read-only location
//(e.g. a cached module checkout)
func (p *Pkg) ReadOnly() bool {

	absPath := p.AbsolutePath()

	modulePath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")

	//TODO: check the file permissions on package files to check

	return strings.Contains(absPath, modulePath)

}

//goPkg validates that the absolutePath denotes a package with at least one go
//file. If there's an error will default to false.
func (g *Pkg) goPkg() bool {

	infos, _ := ioutil.ReadDir(g.AbsolutePath())

	for _, info := range infos {
		if filepath.Ext(info.Name()) == ".go" {
			return true
		}
	}

	return false

}

//Import returns the string that could be used in your source to import this
//package, for exampjle "github.com/jkomoros/boardgame/examples/memory"
func (p *Pkg) Import() (string, error) {
	//Calculate it if not already calculated (for example via NewFromImport constructor)
	if p.importPath == "" {

		goPkg, err := build.ImportDir(p.AbsolutePath(), 0)

		if err != nil {
			return "", errors.New("Couldn't read package: " + err.Error())
		}

		//TODO: factor this into a helper that also sets the package name in
		//case it's asked for later.
		p.importPath = goPkg.ImportPath
	}

	return p.importPath, nil
}

//isPkg verifies that the package appears to be a valid game package.
//Specifically it checks for
func (g *Pkg) isGamePkg() (bool, error) {
	if !g.calculatedIsGamePkg {
		g.memoizedIsGamePkg, g.memoizedIsGamePkgErr = g.calculateIsGamePkg()
	}
	return g.memoizedIsGamePkg, g.memoizedIsGamePkgErr
}

func (g *Pkg) calculateIsGamePkg() (bool, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), g.AbsolutePath(), nil, 0)

	if err != nil {
		return false, errors.New("Couldn't parse folder: " + err.Error())
	}

	if len(pkgs) < 1 {
		return false, errors.New("No packages in that directory")
	}

	if len(pkgs) > 1 {
		return false, errors.New("More than one package in that directory")
	}

	var pkg *ast.Package

	for _, p := range pkgs {
		pkg = p
	}

	foundNewDelegate := false

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fun, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fun.Name.String() != "NewDelegate" {
				continue
			}

			//OK, it might be the function. Does it have the right signature?

			if fun.Recv != nil {
				return false, errors.New("NewDelegate had a receiver")
			}

			if fun.Type.Params.NumFields() > 0 {
				return false, errors.New("NewDelegate took more than 0 items")
			}

			if fun.Type.Results.NumFields() != 1 {
				return false, errors.New("NewDelegate didn't return exactly one item")
			}

			//TODO: check that the returned item implements
			//boardgame.GameDelegate.

			foundNewDelegate = true
			break

		}

		if foundNewDelegate {
			break
		}
	}

	if !foundNewDelegate {
		return false, errors.New("Couldn't find NewDelegate")
	}

	return true, nil
}
