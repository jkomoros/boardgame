package build

import (
	"bytes"
	"errors"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const apiSubFolder = "api"

/*

Api generates the code for a server with the following imported games and
given storage type in a folder called api/ within the given directory, builds
it, and returns the path to the compiled binary. The bulk of the logic to
generate the code is in ApiCode.

To clean up the binary, call CleanupApi and pass the same directory.

*/
func Api(directory string, managers []string, storage StorageType) (string, error) {

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return "", errors.New("The provided directory, " + directory + " does not exist.")
	}

	code, err := ApiCode(managers, storage)

	if err != nil {
		return "", errors.New("Couldn't generate code: " + err.Error())
	}

	apiDir := filepath.Join(directory, apiSubFolder)

	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		if err := os.Mkdir(apiDir, 0700); err != nil {
			return "", errors.New("Couldn't create api directory: " + err.Error())
		}
	}

	codePath := filepath.Join(directory, apiSubFolder, "main.go")

	if err := ioutil.WriteFile(codePath, code, 0644); err != nil {
		return "", errors.New("Couldn't save code: " + err.Error())
	}

	cmd := exec.Command("go", "build")
	cmd.Dir = filepath.Join(directory, apiSubFolder)

	errBuf := new(bytes.Buffer)
	cmd.Stderr = errBuf

	err = cmd.Run()

	if err != nil {
		return "", errors.New("Couldn't build binary: " + err.Error() + ": " + errBuf.String())
	}

	//The binary will have the name of the subfolder it was created in.
	binaryName := filepath.Join(directory, apiSubFolder, apiSubFolder)

	if _, err := os.Stat(binaryName); os.IsNotExist(err) {
		return "", errors.New("Sanity check failed: binary does not appear to have been created.")
	}

	return binaryName, nil
}

//ApiCode returns the code for an api server with the given type.
func ApiCode(managers []string, storage StorageType) ([]byte, error) {

	buf := new(bytes.Buffer)

	managerPkgNames := make([]string, len(managers))

	for i, manager := range managers {
		managerPkgNames[i] = filepath.Base(manager)
	}

	storageImport := storage.Import()

	if storageImport != "" {
		storageImport = "\"" + storageImport + "\""
	}

	err := apiTemplate.Execute(buf, map[string]interface{}{
		"managers":           managers,
		"managerNames":       managerPkgNames,
		"storageImport":      storageImport,
		"storageConstructor": storage.Constructor(),
	})

	if err != nil {
		return nil, errors.New("Couldn't execute code template: " + err.Error())
	}

	formatted, err := format.Source(buf.Bytes())

	if err != nil {
		return nil, errors.New("Couldn't format code output: " + err.Error())
	}

	return formatted, nil

}

//CleanApi removes the api/ directory (code and binary) that was generated
//within directory by ApiCode.
func CleanApi(directory string) error {
	return os.RemoveAll(filepath.Join(directory, apiSubFolder))
}

var apiTemplateText = `/*

A server binary generated automatically by 'boardgame-util/lib/build.Api()'

*/
package main

import (
	{{- range .managers}}
	"{{.}}"
	{{- end}}
	"github.com/jkomoros/boardgame/server/api"
	{{.storageImport}}
)

func main() {

	storage := api.NewServerStorageManager({{.storageConstructor}})
	defer storage.Close()
	api.NewServer(storage,
		{{- range .managerNames}}
		{{.}}.NewDelegate(),
		{{- end}}
	).Start()
}

`
