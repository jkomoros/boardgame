package stub

import (
	"errors"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

//If true, will save out the files generated. Useful for generating new golden
//output when output is changed.
const generateNewGolden = false

const testDir = "test"

func TestGenerate(t *testing.T) {

	tmpls, err := DefaultTemplateSet()

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(tmpls)).DoesNotEqual(0)

	opt := &Options{
		Name: "checkers",
	}

	contents, err := tmpls.Generate(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(contents)).DoesNotEqual(0)

	assert.For(t).ThatActual(contents["checkers/main.go"]).IsNotNil()

	//Now that we unit tested underlying stuff, use Generate() top level,
	//which also formats.

	contents, err = Generate(opt)

	assert.For(t).ThatActual(err).IsNil()

	if generateNewGolden {

		//Save out contents as new golden files to compare against
		contents.Save(testDir, true)
		return
	}

	golden, err := fileContentsFromDir(testDir)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(contents).Equals(golden).ThenDiffOnFail()

}

//fileContentsFromDir loads up filecontents from the given path so they can be
//compared to the golden.
func fileContentsFromDir(path string) (FileContents, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New(path + " doesnt' exist")
	}

	result := make(FileContents)

	if err := recursiveListFilesForFileContents(path, "", result); err != nil {
		return nil, errors.New("couldn't list files: " + err.Error())
	}

	return result, nil

}

//basePath is actual dir to list recursively; prefix is the prefix to affix to
//dir contenst to put in contents.
func recursiveListFilesForFileContents(basePath, prefix string, contents FileContents) error {

	infos, err := ioutil.ReadDir(basePath)

	if err != nil {
		return errors.New("Couldn't list path: " + err.Error())
	}

	for _, info := range infos {
		if info.IsDir() {
			if err := recursiveListFilesForFileContents(filepath.Join(basePath, info.Name()), filepath.Join(prefix, info.Name()), contents); err != nil {
				return err
			}
			continue
		}
		//info represents a file.

		content, err := ioutil.ReadFile(filepath.Join(basePath, info.Name()))

		if err != nil {
			return errors.New("couldn't read " + filepath.Join(basePath, info.Name()) + ": " + err.Error())
		}

		contents[filepath.Join(prefix, info.Name())] = content
	}

	return nil

}
