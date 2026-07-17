/*
Package path includes a few simple convenience methods for dealing with paths
*/
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// AbsoluteGoPkgPath takes a pkg import and returns the full path to the pkg on
// this system. The pkgImport must denote an actual package of go files or it
// will error. It first looks for the right package in $GOPATH, and returns
// that if it finds it. If that doesn't work it falls back on the Go package
// loader, which may download it if it cannot already be satisfied locally. Because
// this uses the $GOPATH copy first, that allows for example relying on games
// locally without going through a VCS, which is nice if you're not connected
// to the internet. If you're trying to load up Game Packages, you should
// likely use the lib/gamepkg package directly.
func AbsoluteGoPkgPath(pkgImport string) (string, error) {
	return AbsoluteGoPkgPathWithOptions(pkgImport, Options{})
}

// Options controls how package paths are resolved.
type Options struct {
	// ReadOnly prevents package loading from modifying go.mod or go.sum.
	ReadOnly bool
}

// AbsoluteGoPkgPathWithOptions is AbsoluteGoPkgPath with explicit resolution
// behavior. ReadOnly is appropriate for validation and CI commands.
func AbsoluteGoPkgPathWithOptions(pkgImport string, options Options) (string, error) {

	//TODO: look into supporting the "no VCS" use case with replace
	//directives, as described here: https://github.com/golang/go/wiki/Modules
	//#what-is-the-status-of-module-support-in-ides-editors-and-standard-
	//tools-like-goimports-gorename-etc

	goPath := os.Getenv("GOPATH")
	if goPath != "" {
		//Check to see if the package at that location exists
		fullPkgPath := filepath.Join(goPath, "src", pkgImport)
		if _, err := os.Stat(fullPkgPath); err == nil {
			return fullPkgPath, nil
		}
	}

	config := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles}
	if options.ReadOnly {
		config.BuildFlags = append(config.BuildFlags, "-mod=readonly")
	}
	loaded, err := packages.Load(config, pkgImport)
	if err != nil {
		return "", fmt.Errorf("load package: %w", err)
	}
	if len(loaded) != 1 {
		return "", fmt.Errorf("package query returned %d packages, want 1", len(loaded))
	}
	if len(loaded[0].Errors) > 0 {
		messages := make([]string, 0, len(loaded[0].Errors))
		for _, problem := range loaded[0].Errors {
			messages = append(messages, problem.Error())
		}
		return "", errors.New(strings.Join(messages, "; "))
	}
	if loaded[0].Dir != "" {
		return loaded[0].Dir, nil
	}
	files := loaded[0].GoFiles
	if len(files) == 0 {
		return "", errors.New("package contained no active Go files")
	}
	return filepath.Dir(files[0]), nil

}

// RelativizePaths takes two absolute paths and returns a string that is the
// relative path from from to to.
func RelativizePaths(from, to string) (string, error) {

	//TODO: pop this out to another more generic place

	if !filepath.IsAbs(from) {
		return "", errors.New("From is not absolute")
	}

	if !filepath.IsAbs(to) {
		return "", errors.New("To is not absolute")
	}

	// filepath.Rel handles roots, volumes, separator normalization, and parent
	// traversal without maintaining a second, subtly different path algorithm.
	// Assembly callers canonicalize symlinked roots before reaching this seam.
	return filepath.Rel(filepath.Clean(from), filepath.Clean(to))

}
