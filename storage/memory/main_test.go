package memory

import (
	"github.com/jkomoros/boardgame/storage/internal/test"
	"testing"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager()
	}, "memory", "", t)

}
