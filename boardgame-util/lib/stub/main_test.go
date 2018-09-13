package stub

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

//If true, will save out the files generated. Useful for generating new golden
//output when output is changed. Flip to true, run `go test`, verify the diff
//looks right, and then flip this back to false before committing.
const generateNewGolden = false

//The go tool will ignore everything rooted in 'testdata'
const testDir = "testdata"

func TestBasicGenerate(t *testing.T) {

	opt := &Options{
		Name: "checkers",
	}

	tmpls, err := DefaultTemplateSet(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(tmpls)).DoesNotEqual(0)

	contents, err := tmpls.Generate(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(contents)).DoesNotEqual(0)

	assert.For(t).ThatActual(contents["checkers/main.go"]).IsNotNil()
}

func TestGolden(t *testing.T) {

	minimalOptions := &Options{
		//ensure we validate name
		Name: " Checkers",
	}

	minimalOptions.SuppressClient()
	minimalOptions.SuppressExtras()

	tutorialOptions := &Options{
		Name:        "checkers",
		DisplayName: "Checkers",
	}

	tutorialOptions.EnableTutorials()

	tests := map[string]*Options{
		"default": {
			Name:              "checkers",
			DisplayName:       "Checkers",
			Description:       "A classic game for two players where you advance across the board, capturing the other player's pawns",
			MinNumPlayers:     2,
			MaxNumPlayers:     4,
			DefaultNumPlayers: 2,
		},
		"minimal":  minimalOptions,
		"tutorial": tutorialOptions,
	}

	if generateNewGolden {
		fmt.Println("Saving new golden. Before committing, flip generateNewGolden back to false.")
	}

	for name, opt := range tests {
		compareGolden(t, name, opt)
	}

}

func compareGolden(t *testing.T, name string, opt *Options) {

	contents, err := Generate(opt)

	assert.For(t, name).ThatActual(err).IsNil()

	dir := filepath.Join(testDir, name)

	if generateNewGolden {

		//Save out contents as new golden files to compare against
		contents.Save(dir, true)

		gameDir := filepath.Join(dir, opt.Name)

		cmd := exec.Command("go", "generate")
		cmd.Dir = gameDir
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			fmt.Println("Couldn't generate: " + err.Error())
			return
		}

		//Generated golden; now verify that the generated pass tests. We do
		//this now so that general tests will be fast; we verify that future
		//tests output the same thing, and then verify that the thing they
		//equal was valid when generated.
		cmd = exec.Command("go", "test")
		cmd.Dir = filepath.Join(dir, opt.Name)
		buf := &bytes.Buffer{}
		cmd.Stderr = buf

		if err := cmd.Run(); err != nil {
			fmt.Println("New package didn't pass test: " + name + ": " + err.Error())
			fmt.Println(buf.String())
			t.FailNow()
			return
		}

		return
	} else if name == "tutorial" {
		//We also do a lot of the expensive building and testing for tutorial,
		//as a tripline to have tests fail when the underlying libraries have
		//changed and the stub outputs need updating.

		tempDir, err := ioutil.TempDir("", "TEMP_test_pkg_")

		if err != nil {
			t.Fatal("Couldn't create temp dir")
		}

		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Fatal("couldn't clean up temp testing dir: " + err.Error())
			}
		}()

		if err := contents.Save(tempDir, false); err != nil {
			t.Error("couldn't save contents: " + err.Error())
		}

		//TODO: this is substantially recreated from right above, which is
		//error-prone.

		gameDir := filepath.Join(tempDir, opt.Name)

		cmd := exec.Command("go", "generate")
		cmd.Dir = gameDir
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			fmt.Println("Couldn't generate: " + err.Error())
			return
		}

		//Generated golden; now verify that the generated pass tests. We do
		//this now so that general tests will be fast; we verify that future
		//tests output the same thing, and then verify that the thing they
		//equal was valid when generated.
		cmd = exec.Command("go", "build")
		cmd.Dir = filepath.Join(tempDir, opt.Name)
		buf := &bytes.Buffer{}
		cmd.Stderr = buf

		if err := cmd.Run(); err != nil {
			t.Fatal("Didn't build (likely underlying library changed) " + err.Error() + ": " + buf.String())
		}

	}

	golden, err := fileContentsFromDir(dir)

	assert.For(t, name).ThatActual(err).IsNil()

	assert.For(t, name).ThatActual(contents).Equals(golden).ThenDiffOnFail()

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

		//Skip auto-generated files
		if strings.HasPrefix(info.Name(), "auto_") && strings.HasSuffix(info.Name(), ".go") {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(basePath, info.Name()))

		if err != nil {
			return errors.New("couldn't read " + filepath.Join(basePath, info.Name()) + ": " + err.Error())
		}

		contents[filepath.Join(prefix, info.Name())] = content
	}

	return nil

}
