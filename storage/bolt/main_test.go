package bolt

import (
	"github.com/jkomoros/boardgame/storage/test"
	"testing"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager(".testdb")
	}, t)

}
