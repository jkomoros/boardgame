/*
	path includes a few simple convenience methods for dealing with paths
*/
package path

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

//AbsoluteGoPkgPath takes a pkg import and returns the full path to the pkg on
//this system. The pkgImport must denote an actual package of go files or it
//will error.
func AbsoluteGoPkgPath(pkgImport string) (string, error) {

	_, err := exec.LookPath("go")

	if err != nil {
		return "", errors.New("go tool not installed")
	}

	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd := exec.Command("go", "list", "-f='{{.Dir}}'", pkgImport)
	cmd.Stdout = buf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		return "", errors.New("go list failed: " + err.Error() + ": " + errBuf.String())
	}

	outputParts := strings.Split(buf.String(), "\n")

	if len(outputParts) < 1 {
		return "", errors.New("No content returned from go list unexpectedly")
	}

	result := outputParts[0]

	result = strings.TrimPrefix(result, "'")
	result = strings.TrimSuffix(result, "'")

	return result, nil

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
