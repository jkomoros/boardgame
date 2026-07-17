/*
Package gamepkg is a package that helps locate, validate, and modify game
package imports.
*/
package gamepkg

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const clientSubFolder = "client"

// RandMagicComment is the string the tool looks for. If this comment is included
// in a source file, then pkg will not error will even if that file does import
// math.Rand(). This comment asserts that the package is using math/rand for some
// reason other than game logic, because game logic is supposed to use
// state.Rand() in order to be predictable.
const RandMagicComment = "boardgame:assert(rand_use_deterministic)"

// Pkg represents a Package that may or may not be a GamePkg.
type Pkg struct {
	//Every contstructo sets absolutePath to something that at least exists on
	//disk.
	absolutePath string
	importPath   string
	name         string
}

// Packages is a convenience func that takes a list of arguments to pass to
// New() (paths or imports) and returns a list of all of the valid packages.
// Any packages that errored for any reason will have their error contained in
// the map of errors. If len(errors) == 0 then no packages errored.
// optionalBasePath will be passed on to New().
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

// AllPackages is a wrapper around Packages that will return a single error and
// no packages if any of the packages was invalid.
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

// New is a wrapper around NewFromImport and NewFromPath. First, it tries to
// interpret the input as an import. If that files, tries to interpret it as a
// path (rel or absolute), and if that fails, bails. optionalBasePath is what
// to pass to NewFromPath if that is used.
func New(importOrPath string, optionalBasePath string) (*Pkg, error) {
	return NewWithOptions(importOrPath, optionalBasePath, Options{})
}

// Options controls package resolution behavior.
type Options struct {
	// ReadOnly prevents package discovery from modifying go.mod or go.sum.
	ReadOnly bool
}

// NewWithOptions is New with explicit package resolution behavior.
func NewWithOptions(importOrPath string, optionalBasePath string, options Options) (*Pkg, error) {
	pathCandidate := importOrPath
	if !filepath.IsAbs(pathCandidate) {
		pathCandidate = filepath.Join(optionalBasePath, pathCandidate)
	}
	if info, err := os.Stat(pathCandidate); err == nil && info.IsDir() {
		return NewFromPathWithOptions(importOrPath, optionalBasePath, options)
	}
	pkg, tryPath, err := newFromImport(importOrPath, options)
	if err == nil {
		return pkg, nil
	}
	if !tryPath {
		return nil, err
	}
	return NewFromPathWithOptions(importOrPath, optionalBasePath, options)
}

// NewFromPath takes path (either relative or absolute path) and returns a new
// Pkg. Will error if the given path does not appear to denote a valid game
// package for any reason. If the path is not absolute, will join wiht
// optionalBasePath (can be either a rel or absolute path). If optionalBasePath
// is "" it will be set to current working directory automatically.
func NewFromPath(path string, optionalBasePath string) (*Pkg, error) {
	return NewFromPathWithOptions(path, optionalBasePath, Options{})
}

// NewFromPathWithOptions is NewFromPath with explicit package resolution
// behavior.
func NewFromPathWithOptions(path string, optionalBasePath string, options Options) (*Pkg, error) {

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

	p, _, e := newPkg(path, "", options)
	return p, e

}

// NewFromImport will return a new Pkg pointing to that import. Will error
// if the given path does not appear to denote a valid game package for any
// reason.
func NewFromImport(importPath string) (*Pkg, error) {
	p, _, e := newFromImport(importPath, Options{})
	return p, e
}

func newFromImport(importPath string, options Options) (pack *Pkg, tryPath bool, err error) {
	analyses, err := Analyze([]string{importPath}, "", options)
	if err != nil {
		return nil, true, errors.New("absolute path couldn't be found: " + err.Error())
	}
	if len(analyses) != 1 {
		return nil, true, errors.New("import resolved to an unexpected number of packages")
	}
	return newPkgFromAnalysis(analyses[0], importPath)
}

// tryPath means, if we fail, should we try using the input as a path?
func newPkg(absPath, importPath string, options Options) (p *Pkg, tryPath bool, err error) {

	if info, err := os.Stat(absPath); err != nil {
		return nil, true, errors.New("Path doesn't point to valid location on disk: " + err.Error())
	} else if !info.IsDir() {
		return nil, true, errors.New("path points to an object but it's not a directory")
	}

	analyses, err := Analyze([]string{"."}, absPath, options)
	if err != nil {
		return nil, true, errors.New(absPath + " could not be analyzed: " + err.Error())
	}
	if len(analyses) != 1 {
		return nil, true, errors.New(absPath + " resolved to an unexpected number of packages")
	}
	return newPkgFromAnalysis(analyses[0], importPath)
}

func newPkgFromAnalysis(analysis Analysis, providedImport string) (p *Pkg, tryPath bool, err error) {
	if !analysis.Candidate {
		return nil, true, &InvalidGamePackageError{Analysis: analysis}
	}
	if len(analysis.Diagnostics) > 0 {
		return nil, false, &InvalidGamePackageError{Analysis: analysis}
	}

	if providedImport != "" {
		if providedImport != analysis.ImportPath {
			return nil, true, errors.New("the provided import path does not agree with typed package loading: " + providedImport + " : " + analysis.ImportPath)
		}
	}
	return &Pkg{absolutePath: analysis.Dir, importPath: analysis.ImportPath, name: analysis.Name}, true, nil
}

func formatDiagnostics(diagnostics []Diagnostic) string {
	parts := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		location := diagnostic.Position.File
		if diagnostic.Position.Line > 0 {
			location += ":" + strconv.Itoa(diagnostic.Position.Line)
			if diagnostic.Position.Column > 0 {
				location += ":" + strconv.Itoa(diagnostic.Position.Column)
			}
		}
		if location != "" {
			parts = append(parts, location+": "+diagnostic.Message)
		} else {
			parts = append(parts, diagnostic.Message)
		}
	}
	return strings.Join(parts, "; ")
}

// AbsolutePath returns the absolute path where the package in question resides
// on disk. All constructors will have errored if AbsolutePath doesn't at the
// very least point to a valid location on disk. For example, "/Users/YOURUSERNAME/Code/go/src/github.com/jkomoros/boardgame/examples/memory"
func (p *Pkg) AbsolutePath() string {
	return p.absolutePath
}

// ReadOnly returns true if the package appears to be in a read-only location
// (e.g. a cached module checkout)
func (p *Pkg) ReadOnly() bool {

	absPath := p.AbsolutePath()

	modulePath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")

	//TODO: check the file permissions on package files to check

	return strings.Contains(absPath, modulePath)

}

// EnsureDir ensures the given directory, relative to package root, exists.
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

// WriteFile writes the given relPath contents with 0644 perms. If overwite is
// true will overwrite; if overwrite is false and the file already exists will
// fail.
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

// RemoveFile removes the given path, relative to the base of the package, from
// the package if possible.
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

// RemoveDirIfEmpty removes the given dir if it contains no items.
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

// ClientFolder returns the absolute path to this game package's folder of
// client assets, or "" if this game does not have a client folder. Example: "/Users/YOURUSERNAME/Code/go/src/github.com/jkomoros/boardgame/examples/memory/client"
func (p *Pkg) ClientFolder() string {
	path := filepath.Join(p.AbsolutePath(), clientSubFolder)
	if p.Has(clientSubFolder) {
		return path
	}
	return ""
}

// Has returns whether the given relPath (directory or file) exists relative to
// this package.
func (p *Pkg) Has(relPath string) bool {
	path := filepath.Join(p.AbsolutePath(), relPath)

	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

// Import returns the string that could be used in your source to import this
// package, for exampjle "github.com/jkomoros/boardgame/examples/memory"
func (p *Pkg) Import() string {

	return p.importPath
}

// Name returns the package name, according to a static analysis of the source.
// Technically it's possible that this differs from the package's delegate's
// Name(), however in practice that's extremely unlikely because the core
// library will fail to create a GameManager if the package and delegate name
// don't match. That means that the return value of this method can effectively
// be used as though it equals the delegate's Name().
func (p *Pkg) Name() string {
	return p.name
}
