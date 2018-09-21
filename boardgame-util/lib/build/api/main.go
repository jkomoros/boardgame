package api

import (
	"bytes"
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const apiSubFolder = "api"

//Options is a struct to pass extra options to Code() and Build(). The
//defaults are all the zero values.
type Options struct {
	//If true, installs an overrider in the generated binary that enables
	//offline dev mode.
	OverrideOfflineDevMode bool
}

/*

Build is the primary method in this package. It generates the code for a
server with the following imported games and given storage type in a folder
called api/ within the given directory, builds it, and returns the path to the
compiled binary. The bulk of the logic to generate the code is in Code().

To clean up the binary, call Cleanup and pass the same directory.

*/
func Build(directory string, pkgs []*gamepkg.Pkg, storage StorageType, options *Options) (string, error) {

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return "", errors.New("The provided directory, " + directory + " does not exist.")
	}

	code, err := Code(pkgs, storage, options)

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

//Code returns the code for the `api/main.go`of a server with the given type.
//Options may be nil for default options.
func Code(pkgs []*gamepkg.Pkg, storage StorageType, options *Options) ([]byte, error) {

	if options == nil {
		options = &Options{}
	}

	buf := new(bytes.Buffer)

	storageImport := storage.Import()

	if storageImport != "" {
		storageImport = "\"" + storageImport + "\""
	}

	err := apiTemplate.Execute(buf, map[string]interface{}{
		"pkgs":               pkgs,
		"storageImport":      storageImport,
		"storageConstructor": storage.Constructor(),
		"options":            options,
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

//Clean removes the api/ directory (code and binary) that was generated
//within directory by Build.
func Clean(directory string) error {
	return os.RemoveAll(filepath.Join(directory, apiSubFolder))
}

var apiTemplateText = `/*

A server binary generated automatically by 'boardgame-util/lib/build/api/Build()'

*/
package main

import (
	{{- range .pkgs}}
	"{{.Import}}"
	{{- end}}
	"github.com/jkomoros/boardgame/server/api"
	{{.storageImport}}
	{{- if .options.OverrideOfflineDevMode }}
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	{{- end}}
)

{{if .options.OverrideOfflineDevMode}}
var overrides []config.OptionOverrider

func init() {
	overrides = append(overrides, config.EnableOfflineDevMode())
}
{{end}}

func main() {

	storage := api.NewServerStorageManager({{.storageConstructor}})
	defer storage.Close()
	api.NewServer(storage,
		{{- range .pkgs}}
		{{.Name}}.NewDelegate(),
		{{- end}}
	{{- if .options.OverrideOfflineDevMode }}		
	).AddOverrides(overrides).Start()
	{{- else}}
	).Start()
	{{- end}}
}

`
