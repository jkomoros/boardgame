package memory

import (
	"testing"

	"github.com/jkomoros/boardgame/storage/internal/test"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager()
	}, "memory", "", t)

}
