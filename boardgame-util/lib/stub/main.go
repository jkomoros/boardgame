/*

	stub is a library that helps generate stub code for new games

*/
package stub

import (
	"errors"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	//The name of the game
	Name string
}

//FileContents is the generated contents of the files to later write to the
//filesystem.
type FileContents map[string][]byte

//Generate generates FileContents for the given set of options.
func Generate(opt *Options) (FileContents, error) {

	//TODO: gofmt the code before returning
	return nil, errors.New("Not yet implemented")
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
//fine. Will error if any of the files to create already exist. After saving,
//runs `go generate`.
func (f FileContents) Save(dir string) error {

	for name := range f {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return errors.New(name + " already existed; save aborted")
		}
	}

	return errors.New("Not yet implemented")
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
