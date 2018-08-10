/*

build is a package that can create and cleanup api server binaries, static
asset folders, and golden test setups.

Typically it is not used directly, but via the `boardgame-util build` and
`boardgame-util cleanup` commands.

*/
package build

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type StorageType int

const (
	StorageInvalid StorageType = iota
	StorageMemory
	StorageBolt
	StorageMysql
	StorageFilesystem
)

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

	return s.String() + "(" + args + ")"

}

/*

Api generates the code for a server with the following imported games and
given storage type, builds it, and returns the path to the compiled binary.

To clean up the binary, call CleanupApi and pass the same directory.

*/
func Api(directory string, managers []string, storage StorageType) (string, error) {
	return "", errors.New("Not yet implemented")
}

func CleanupApi(directory string) error {
	return os.RemoveAll(directory)
}
