/*

Package api is a package that can create and cleanup api server binaries. Its
cousin is the build/static package, which contains considerably more logic.

The point of this package is primarily to create an `api/main.go` output in the
given directory that statically links the games and storage type configured via
config.json, and then build that binary using `go build`.

The directory parameter gives the build directory; the build command will create
an `api` sub-folder within that, and static.Build() will create a static
directory. A directory of "" is legal and is effectively ".".

There's nothing magic about this package; it's legal to create your own server
binary by hand. This package just automates that for you so when you add a game
to your server you only have to worry about adding it in your config.json and
everything else happens automatically.

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

//StorageType denotes one of the storage managers this package knows how to
//generate code for.
type StorageType int

const (
	//StorageInvalid is the invalid default
	StorageInvalid StorageType = iota
	//StorageMemory denotes the memory storage layer
	StorageMemory
	//StorageBolt denotes the bolt storage layer
	StorageBolt
	//StorageMysql denotes the mysql storage layer
	StorageMysql
	//StorageFilesystem denotes the filesystem storage layer
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
		StorageMemory.String(),
		StorageBolt.String(),
		StorageMysql.String(),
		StorageFilesystem.String(),
	}
}

//StorageTypeFromString returns the right storage type for the given string.
//"" returns StorageDefault, and any unknown types return StorageInvalid.
func StorageTypeFromString(in string) StorageType {
	in = strings.ToLower(in)
	in = strings.TrimSpace(in)

	switch in {
	case "":
		return StorageBolt
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

//Import is the string denting the import path for this storage type, e.g.
//"github.com/jkomoros/boardgame/storage/mysql"
func (s StorageType) Import() string {

	base := "github.com/jkomoros/boardgame/storage"
	return filepath.Join(base, s.String())
}

//Constructor is a string representing a default constructor for this storage
//type, e.g. `bolt.NewStorageManager(".database")`. optionalLiteralArgs will
//be passed literally within the `()` of the storage constructor, so valid
//strings are "\".database\"" etc. If not provided, will fall back on
//reasonable defaults for that type.
func (s StorageType) Constructor(optionalLiteralArgs string) string {

	args := ""

	switch s {
	case StorageFilesystem:
		args = "\"games/\""
	case StorageBolt:
		args = "\".database\""
	case StorageMysql:
		args = "false"
	}

	if optionalLiteralArgs != "" {
		args = optionalLiteralArgs
	}

	return s.String() + ".NewStorageManager(" + args + ")"

}
