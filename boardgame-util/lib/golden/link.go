package golden

import (
	"bytes"
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
)

//Path relative to package root where goldens should be stored
const GameRecordsFolder = "testdata/golden"

//The name of the test file to create
const GoldenTestFile = "golden_test.go"

var goldenTestTemplate *template.Template

func init() {
	goldenTestTemplate = template.Must(template.New("golden").Parse(goldenTestTemplateText))
}

//MakeGoldenTest ensures that GameRecordFolder exists and creates
//golden_test.go in the root of the package if it doesn't yet exist.
func MakeGoldenTest(pkg *gamepkg.Pkg) error {

	if pkg == nil {
		return errors.New("No package provided")
	}

	if err := pkg.EnsureDir(GameRecordsFolder); err != nil {
		return errors.New("Couldn't ensure game records dir: " + err.Error())
	}

	buf := new(bytes.Buffer)

	err := goldenTestTemplate.Execute(buf, map[string]string{
		"gametype": pkg.Name(),
		"folder":   GameRecordsFolder,
	})

	if err != nil {
		return errors.New("Couldn't generate blob from template: " + err.Error())
	}

	return pkg.WriteFile(GoldenTestFile, buf.Bytes(), true)

}

//LinkGoldenFolders helps create a folder system for the `filesystem` storage
//layer, with the individual folders pointing back to folders adjacent to the
//games they're affiliated with. This is useful when you want to generate new
//golden tests for game types. First we use reflection to find the package
//path for each delegate, ensure a folder exists within it with tbe
//goldenFolderName name, create a soft-link from basePath to that folder, and
//create a `golden_test.go` file that automatically tests all of those golden
//files (and assumes that your package defines a `NewDelegate()
//boardgame.GameDelegate` method). The result is that the underlying files
//will be stored in folders adjacent to the games they are relative to, which
//is convenient if you're adding new golden games to the test set. The
//`filesystem` storage layer will call this as a convenience if you pass a
//non-"" goldenFolderName to filesystem.NewStorageManager.
func LinkGoldenFolders(basePath string, goldenFolderName string, managers []*boardgame.GameManager) error {

	if goldenFolderName == "" {
		goldenFolderName = "golden"
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return errors.New(basePath + " does not already exist.")
	}

	for _, manager := range managers {
		pkgPath := reflect.ValueOf(manager.Delegate()).Elem().Type().PkgPath()
		if err := linkGoldenFolder(manager.Delegate().Name(), pkgPath, basePath, goldenFolderName); err != nil {
			return errors.New("Couldn't link golden folder for " + manager.Delegate().Name() + ": " + err.Error())
		}
	}

	return nil
}

func linkGoldenFolder(gameType, pkgPath, basePath, goldenFolderName string) error {

	//TODO: should this be public?

	//This SHOULD handle vendored games correctly, given that
	//reflect.PkgPath() returns using the full path, including /vendor/

	fullPkgPath, err := path.AbsoluteGoPkgPath(pkgPath)

	if err != nil {
		return errors.New("Couldn't get full pkg path: " + err.Error())
	}

	fullPath := filepath.Join(fullPkgPath, goldenFolderName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Println("Creating " + fullPath)
		if err := os.Mkdir(fullPath, 0700); err != nil {
			return errors.New("Could not make golden path: " + err.Error())
		}
	}

	gamePath := filepath.Join(basePath, gameType)

	if _, err := os.Stat(gamePath); os.IsNotExist(err) {

		execPath, err := os.Executable()

		if err != nil {
			return errors.New("Couldn't get executable path: " + err.Error())
		}

		relPath, err := path.RelativizePaths(execPath, fullPath)

		if err != nil {
			return errors.New("Couldn't relativize paths: " + err.Error())
		}

		log.Println("Linking " + gamePath + " to " + relPath)

		//Soft link from basePath.

		if err := os.Symlink(relPath, gamePath); err != nil {
			return errors.New("Couldn't create symlink: " + err.Error())
		}
	}

	//TODO: allow this to be skipped as an otpion.
	if err := createGoldenTest(fullPkgPath, goldenFolderName); err != nil {
		return errors.New("Couldn't store golden test: " + err.Error())
	}

	return nil

}

func createGoldenTest(fullPkgPath, goldenFolderName string) error {

	//TODO: should this be public?

	pkgName, err := verifyPkgForGolden(fullPkgPath)

	if err != nil {
		return errors.New("Package didn't validate: " + err.Error())
	}

	buf := new(bytes.Buffer)

	err = goldenTestTemplate.Execute(buf, map[string]string{
		"gametype": pkgName,
		"folder":   goldenFolderName,
	})

	if err != nil {
		return errors.New("Couldn't generate blob from template: " + err.Error())
	}

	return ioutil.WriteFile(filepath.Join(fullPkgPath, "golden_test.go"), buf.Bytes(), 0644)

}

//verifyPkgForGolden looks at the given package, returns the package name, and
//verifies that it has a NewDelegate method.
func verifyPkgForGolden(fullPkgName string) (string, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), fullPkgName, nil, 0)

	if err != nil {
		return "", errors.New("Couldn't parse folder: " + err.Error())
	}

	if len(pkgs) < 1 {
		return "", errors.New("No packages in that directory")
	}

	if len(pkgs) > 1 {
		return "", errors.New("More than one package in that directory")
	}

	var pkg *ast.Package
	pkgName := ""

	for key, p := range pkgs {
		pkgName = key
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
				return "", errors.New("NewDelegate had a receiver")
			}

			if fun.Type.Params.NumFields() > 0 {
				return "", errors.New("NewDelegate took more than 0 items")
			}

			if fun.Type.Results.NumFields() != 1 {
				return "", errors.New("NewDelegate didn't return exactly one item")
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
		return "", errors.New("Couldn't find NewDelegate")
	}

	return pkgName, nil

}

var goldenTestTemplateText = `package {{.gametype}}

/*

This file was created automatically by the filesystem storage layer with a
golden folder.

It will be overwritten the next time a filesystem is booted that uses this
game package.

*/

import (
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"testing"
)

func TestGolden(t *testing.T) {
	if err := golden.CompareFolder(NewDelegate(), "{{.folder}}"); err != nil {
		t.Error(err.Error())
	}
}
`
