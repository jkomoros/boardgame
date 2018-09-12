/*

api is a package that can create and cleanup api server binaries. Its cousin
is the build/static package, which contains considerably more logic.

The point of this package is primarily to create an `api/main.go` output in
the given directory that statically links the games and storage type
configured via config.json, and then build that binary using `go build`.

The directory parameter gives the build directory; the build command will
create an `api` sub-folder within that, and static.Build() will create a
static directory. A directory of "" is legal and is effectively ".".

There's nothing magic about this package; it's legal to create your own server
binary by hand. This package just automates that for you so when you add a
game to your server you only have to worry about adding it in your config.json
and everything else happens automatically.

For a config json that has a defaultstoragetype of bolt and lists the games
`github.com/jkomoros/boardgame/examples/checkers`,
`github.com/jkomoros/boardgame/examples/memory`, and
`github.com/jkomoros/boardgame/examples/pig` it would output (with the package
doc comment omitted):

	package main

	import (
		"github.com/jkomoros/boardgame/examples/checkers"
		"github.com/jkomoros/boardgame/examples/memory"
		"github.com/jkomoros/boardgame/examples/pig"
		"github.com/jkomoros/boardgame/server/api"
		"github.com/jkomoros/boardgame/storage/bolt"
	)

	func main() {

		storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
		defer storage.Close()
		api.NewServer(storage,
			checkers.NewDelegate(),
			memory.NewDelegate(),
			pig.NewDelegate(),
		).Start()
	}

Typically it is not used directly, but via the `boardgame-util build api
`,`boardgame-util cleanup api`, and `boardgame-util serve` commands.

*/
package api

import (
	"path/filepath"
	"strings"
	"text/template"
)

type StorageType int

const (
	StorageInvalid StorageType = iota
	StorageDefault
	StorageMemory
	StorageBolt
	StorageMysql
	StorageFilesystem
)

var apiTemplate *template.Template

func init() {
	apiTemplate = template.Must(template.New("api").Parse(apiTemplateText))
}

//ValidStorageTypeStrings returns an array of strings that are the normal
//(i.e. not invalid) strings that would return useful values if passed to
//StorageTypeFromString.
func ValidStorageTypeStrings() []string {
	return []string{
		StorageDefault.String(),
		StorageMemory.String(),
		StorageBolt.String(),
		StorageMysql.String(),
		StorageFilesystem.String(),
	}
}

func StorageTypeFromString(in string) StorageType {
	in = strings.ToLower(in)
	in = strings.TrimSpace(in)

	switch in {
	case "default":
		return StorageDefault
	case "":
		return StorageDefault
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
	case StorageDefault:
		return "default"
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

	if s == StorageDefault {
		//api package already imported
		return ""
	}

	base := "github.com/jkomoros/boardgame/storage"
	return filepath.Join(base, s.String())
}

//Constructor is a string representing a default constructor for this storage
//type, e.g. `bolt.NewStorageManager(".database")`
func (s StorageType) Constructor() string {

	if s == StorageDefault {
		return "api.NewDefaultStorageManager()"
	}

	args := ""

	switch s {
	case StorageFilesystem:
		args = "\"games/\""
	case StorageBolt:
		args = "\".database\""
	case StorageMysql:
		args = "false"
	}

	return s.String() + ".NewStorageManager(" + args + ")"

}
