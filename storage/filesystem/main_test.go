package filesystem

import (
	"github.com/jkomoros/boardgame/storage/internal/test"
	"testing"
)

func TestStorageManagerDiffEncoding(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager("test")
	}, "filesystem", "", t)

}

func TestStorageManagerFullEncoding(t *testing.T) {

	test.Test(func() test.StorageManager {
		mgr := NewStorageManager("test")
		mgr.forceFullEncoding = true
		return mgr
	}, "filesystem", "", t)

}
