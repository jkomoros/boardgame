/*

build is a package that can create and cleanup api server binaries, static
asset folders, and golden test setups.

Typically it is not used directly, but via the `boardgame-util build` and
`boardgame-util cleanup` commands.

*/
package build

import (
	"bytes"
	"errors"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type StorageType int

const (
	StorageInvalid StorageType = iota
	StorageMemory
	StorageBolt
	StorageMysql
	StorageFilesystem
)

var apiTemplate *template.Template

func init() {
	apiTemplate = template.Must(template.New("api").Parse(apiTemplateText))
}

func StorageTypeFromString(in string) StorageType {
	in = strings.ToLower(in)
	in = strings.TrimSpace(in)

	switch in {
	case "memory":
		return StorageMemory
	case "bolt":
		return StorageBolt
	case "mysql":
		return StorageMysql
	case "filesystem":
		return StorageFilesystem
	}

	return StorageInvalid
}

func (s StorageType) String() string {
	switch s {
	case StorageMemory:
		return "memory"
	case StorageBolt:
		return "bolt"
	case StorageMysql:
		return "mysql"
	case StorageFilesystem:
		return "filesystem"
	}
	return "invalid"
}

//Import is the string denting the import path for this storage type.
func (s StorageType) Import() string {
	base := "github.com/jkomoros/boardgame/storage"
	return filepath.Join(base, s.String())
}

//Constructor is a string representing a default constructor for this storage
//type, e.g. `bolt.NewStorageManager(".database")`
func (s StorageType) Constructor() string {

	args := ""

	switch s {
	case StorageFilesystem:
		args = "games/"
	case StorageBolt:
		args = ".database"
	}

	return s.String() + ".NewStorageManager(\"" + args + "\")"

}

const subFolder = "api/"

/*

Api generates the code for a server with the following imported games and
given storage type in a folder called api/ within the given directory, builds
it, and returns the path to the compiled binary. The bulk of the logic to
generate the code is in ApiCode.

To clean up the binary, call CleanupApi and pass the same directory.

*/
func Api(directory string, managers []string, storage StorageType) (string, error) {
	return "", errors.New("Not yet implemented")
}

//ApiCode returns the code for an api server with the given type.
func ApiCode(managers []string, storage StorageType) ([]byte, error) {

	buf := new(bytes.Buffer)

	managerPkgNames := make([]string, len(managers))

	for i, manager := range managers {
		managerPkgNames[i] = filepath.Base(manager)
	}

	err := apiTemplate.Execute(buf, map[string]interface{}{
		"managers":           managers,
		"managerNames":       managerPkgNames,
		"storageImport":      storage.Import(),
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

func CleanupApi(directory string) error {
	return os.RemoveAll(filepath.Join(directory, subFolder))
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
	"{{.storageImport}}"
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
