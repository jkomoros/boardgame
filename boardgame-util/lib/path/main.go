/*
	path includes a few simple convenience methods for dealing with paths
*/
package path

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

//AbsoluteGoPkgPath takes a pkg import and returns the full path to the pkg on
//this system.
func AbsoluteGoPkgPath(pkgImport string) (string, error) {
	goPath := os.Getenv("GOPATH")

	if goPath == "" {
		return "", errors.New("Gopath wasn't set")
	}

	fullPkgPath := filepath.Join(goPath, "src", pkgImport)

	return fullPkgPath, nil
}

//RelativizePaths takes two absolute paths and returns a string that is the
//relative path from from to to.
func RelativizePaths(from, to string) (string, error) {

	//TODO: pop this out to another more generic place

	if !filepath.IsAbs(from) {
		return "", errors.New("From is not absolute")
	}

	if !filepath.IsAbs(to) {
		return "", errors.New("To is not absolute")
	}

	from = filepath.Clean(from)
	to = filepath.Clean(to)

	prefix := pathPrefix(from, to)

	if prefix == "" {
		return "", errors.New("No prefix in common")
	}

	fromRest := strings.TrimPrefix(from, prefix)
	toRest := strings.TrimPrefix(to, prefix)

	fromPieces := strings.Split(fromRest, string(filepath.Separator))

	dots := make([]string, len(fromPieces))

	for i := range fromPieces {
		dots[i] = ".."
	}

	return filepath.Join(filepath.Join(dots...), toRest), nil

}

func pathPrefix(from, to string) string {

	var overlappingParts []string

	fromParts := strings.Split(from, string(filepath.Separator))
	toParts := strings.Split(to, string(filepath.Separator))

	for i, fromPart := range fromParts {
		if i >= len(toParts) {
			break
		}
		toPart := toParts[i]

		if fromPart != toPart {
			break
		}
		overlappingParts = append(overlappingParts, fromPart)
	}

	return strings.Join(overlappingParts, string(filepath.Separator)) + string(filepath.Separator)

}
