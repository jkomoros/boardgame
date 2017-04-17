package mysql

import (
	"github.com/jkomoros/boardgame/storage/test"
	"testing"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager(true)
	}, "mysql", "root:root@(localhost:3306)/boardgame-test", t)

}
