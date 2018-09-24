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

const clientSubFolder = "client"

type Pkg struct {
	//Every contstructo sets absolutePath to something that at least exists on
	//disk.
	absolutePath         string
	importPath           string
	name                 string
	calculatedIsGamePkg  bool
	memoizedIsGamePkg    bool
	memoizedIsGamePkgErr error
}

//Packages is a convenience func that takes a list of arguments to pass to
//New() (paths or imports) and returns a list of all of the valid packages.
//Any packages that errored for any reason will have their error contained in
//the map of errors. If len(errors) == 0 then no packages errored.
//optionalBasePath will be passed on to New().
func Packages(inputs []string, optionalBasePath string) ([]*Pkg, map[string]error) {
	var result []*Pkg
	errs := make(map[string]error)

	for _, input := range inputs {
		pkg, err := New(input, optionalBasePath)
		if err == nil {
			result = append(result, pkg)
		} else {
			errs[input] = err
		}
	}

	if len(errs) == 0 {
		errs = nil
	}

	return result, errs
}

//AllPackages is a wrapper around Packages that will return a single error and
//no packages if any of the packages was invalid.
func AllPackages(inputs []string, optionalBasePath string) ([]*Pkg, error) {
	pkgs, errs := Packages(inputs, optionalBasePath)

	if len(errs) == 0 {
		return pkgs, nil
	}

	var errorStrings []string
	for key, val := range errs {
		errorStrings = append(errorStrings, key+": "+val.Error())
	}

	return nil, errors.New("At least one package failed to load: " + strings.Join(errorStrings, "; "))
}

//New is a wrapper around NewFromImport and NewFromPath. First, it tries to
//interpret the input as an import. If that files, tries to interpret it as a
//path (rel or absolute), and if that fails, bails. optionalBasePath is what
//to pass to NewFromPath if that is used.
func New(importOrPath string, optionalBasePath string) (*Pkg, error) {
	pkg, err := NewFromImport(importOrPath)
	if err == nil {
		return pkg, nil
	}
	return NewFromPath(importOrPath, optionalBasePath)
}

//NewFromPath takes path (either relative or absolute path) and returns a new
//Pkg. Will error if the given path does not appear to denote a valid game
//package for any reason. If the path is not absolute, will join wiht
//optionalBasePath (can be either a rel or absolute path). If optionalBasePath
//is "" it will be set to current working directory automatically.
func NewFromPath(path string, optionalBasePath string) (*Pkg, error) {

	if !filepath.IsAbs(path) {

		//If optionalBasePath is "" this is a no op
		path = filepath.Join(optionalBasePath, path)

		//if it's still not absolute then optionalBasePath must have been "" or a rel path itself.
		if !filepath.IsAbs(path) {
			cwd, err := os.Getwd()

			if err != nil {
				return nil, errors.New("Couldn't get working directory: " + err.Error())
			}

			path = filepath.Join(cwd, path)
		}
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

	//We also ensure we have a good value for importPath now, so that Import()
	//later can just return a string, not (string, error)

	goPkg, err := build.ImportDir(absPath, 0)

	if err != nil {
		return nil, errors.New("Couldn't read package: " + err.Error())
	}

	if importPath != "" {
		if importPath != goPkg.ImportPath {
			return nil, errors.New("The provided import path does not agree with what go.build thinks the import path is: " + importPath + " : " + goPkg.ImportPath)
		}
	}
	result.importPath = goPkg.ImportPath
	result.name = goPkg.Name

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

//EnsureDir ensures the given directory, relative to package root, exists.
func (p *Pkg) EnsureDir(relPath string) error {

	dir := filepath.Join(p.AbsolutePath(), relPath)

	if info, err := os.Stat(dir); err == nil {
		if info.IsDir() {
			return nil
		}
		return errors.New("relPath " + relPath + " exists but is not a directory")
	}

	//Need to create it.
	if p.ReadOnly() {
		return errors.New(relPath + " didn't exist, but package was read only")
	}

	return os.MkdirAll(dir, 0700)

}

//WriteFile writes the given relPath contents with 0644 perms. If overwite is
//true will overwrite; if overwrite is false and the file already exists will
//fail.
func (p *Pkg) WriteFile(relPath string, contents []byte, overwrite bool) error {
	if p.ReadOnly() {
		return errors.New("Package is readonly")
	}

	path := filepath.Join(p.AbsolutePath(), relPath)
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return errors.New(relPath + " already existed and overwrite wasn't true")
		}
	}

	return ioutil.WriteFile(path, contents, 0644)

}

func (p *Pkg) RemoveFile(relPath string) error {
	if p.ReadOnly() {
		return errors.New("Package is readonly")
	}
	if !p.Has(relPath) {
		return nil
	}
	path := filepath.Join(p.AbsolutePath(), relPath)
	return os.Remove(path)
}

//RemoveDirIfEmpty removes the given dir if it contains no items.
func (p *Pkg) RemoveDirIfEmpty(relPath string) error {
	if !p.Has(relPath) {
		return nil
	}

	dir := filepath.Join(p.AbsolutePath(), relPath)
	infos, err := ioutil.ReadDir(dir)

	if err != nil {
		return errors.New("Couldn't read dir: " + err.Error())
	}

	if len(infos) != 0 {
		//Items so don't remove
		return nil
	}

	if p.ReadOnly() {
		return errors.New("Package is read only")
	}

	return os.Remove(dir)
}

//ClientFolder returns the absolute path to this game package's folder of
//client assets, or "" if this game does not have a client folder. Example: "/Users/YOURUSERNAME/Code/go/src/github.com/jkomoros/boardgame/examples/memory/client"
func (p *Pkg) ClientFolder() string {
	path := filepath.Join(p.AbsolutePath(), clientSubFolder)
	if p.Has(clientSubFolder) {
		return path
	}
	return ""
}

//Has returns whether the given relPath (directory or file) exists relative to
//this package.
func (p *Pkg) Has(relPath string) bool {
	path := filepath.Join(p.AbsolutePath(), relPath)

	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
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
func (p *Pkg) Import() string {

	return p.importPath
}

//Name returns the package name, which should also be the name of the game
//(what NewDelegate.Name() returns).
func (p *Pkg) Name() string {
	return p.name
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
