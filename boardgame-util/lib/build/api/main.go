package api

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	buildstatic "github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

const apiSubFolder = "api"

// Options is a struct to pass extra options to Code() and Build(). The
// defaults are all the zero values.
type Options struct {
	//If true, installs an overrider in the generated binary that enables
	//offline dev mode.
	OverrideOfflineDevMode bool

	// When non-empty, overrides the generated server's configured CORS and
	// WebSocket origin allowlist without persisting a config-file mutation.
	OverrideAllowedOrigins string

	//Will be passed to the storage's Constructor() method as the
	//optionalLiteralArgs.
	StorageLiteralArgs string
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
		return "", errors.New("sanity check failed: binary does not appear to have been created")
	}

	return binaryName, nil
}

// Code returns the code for the `api/main.go`of a server with the given type.
// Options may be nil for default options.
func Code(pkgs []*gamepkg.Pkg, storage StorageType, options *Options) ([]byte, error) {

	if options == nil {
		options = &Options{}
	}

	buf := new(bytes.Buffer)

	storageImport := storage.Import()
	storageConstructor := storage.Constructor(options.StorageLiteralArgs)

	if storageImport != "" {
		storageAlias := availableStorageAlias(pkgs)
		storageImport = storageAlias + " \"" + storageImport + "\""
		storageConstructor = strings.Replace(storageConstructor, storage.String()+".", storageAlias+".", 1)
	}

	err := apiTemplate.Execute(buf, map[string]interface{}{
		"pkgs":                   pkgs,
		"storageImport":          storageImport,
		"storageConstructor":     storageConstructor,
		"options":                options,
		"hasOverrides":           options.OverrideOfflineDevMode || options.OverrideAllowedOrigins != "",
		"overrideAllowedOrigins": strconv.Quote(options.OverrideAllowedOrigins),
		"companionCapableGames":  buildstatic.CompanionCapableGames(pkgs),
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

func availableStorageAlias(pkgs []*gamepkg.Pkg) string {
	used := map[string]bool{"api": true}
	for _, pkg := range pkgs {
		used[pkg.Name()] = true
	}
	return availableImportAlias(used)
}

func availableImportAlias(used map[string]bool) string {
	for suffix := 0; ; suffix++ {
		candidate := "boardgamestorage"
		if suffix > 0 {
			candidate += fmt.Sprint(suffix)
		}
		if !used[candidate] {
			return candidate
		}
	}
}

// Clean removes the api/ directory (code and binary) that was generated
// within directory by Build.
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
	{{- if .hasOverrides }}
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	{{- end}}
)

{{if .hasOverrides}}
var overrides []config.OptionOverrider

func init() {
	{{- if .options.OverrideOfflineDevMode }}
	overrides = append(overrides, config.EnableOfflineDevMode())
	{{- end}}
	{{- if .options.OverrideAllowedOrigins }}
	overrides = append(overrides, config.OverrideAllowedOrigins({{.overrideAllowedOrigins}}))
	{{- end}}
}
{{end}}

// companionCapableGames is the list of game names that ship the Table+Hand
// renderer pair (boardgame-render-game-<name>-table.ts AND -hand.ts) as of
// this build. Computed by boardgame-util at build time via a filesystem
// walk (see boardgame-util/lib/build/static.CompanionCapableGames). The
// server uses this to populate managerInfo.supportsTableHandMode and surface
// it in doListManager so the create-game form can show the
// "Use shared projector + phones" toggle for supporting games. (Spec §5.3.)
var companionCapableGames = []string{
{{- range .companionCapableGames}}
	"{{.}}",
{{- end}}
}

func main() {

	storage := api.NewServerStorageManager({{.storageConstructor}})
	defer storage.Close()
	api.NewServer(storage,
		{{- range .pkgs}}
		{{.Name}}.NewDelegate(),
		{{- end}}
	{{- if .hasOverrides }}
	).AddOverrides(overrides).WithCompanionCapableGames(companionCapableGames).Start()
	{{- else}}
	).WithCompanionCapableGames(companionCapableGames).Start()
	{{- end}}
}

`
