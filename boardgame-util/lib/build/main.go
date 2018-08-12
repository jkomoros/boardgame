/*

build is a package that can create and cleanup api server binaries, static
asset folders, and golden test setups.

Typically it is not used directly, but via the `boardgame-util build` and
`boardgame-util cleanup` commands.

*/
package build

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
