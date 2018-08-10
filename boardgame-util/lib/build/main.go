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
