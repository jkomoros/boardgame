package filesystem

import (
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/jkomoros/boardgame/storage/internal/test"
	"testing"
)

func TestStorageManagerYudaiEncoding(t *testing.T) {

	test.Test(func() test.StorageManager {
		mgr := NewStorageManager("test")
		mgr.SetStateEncoding(record.StateEncodingYudai)
		return mgr
	}, "filesystem", "", t)

}

func TestStorageManagerFullEncoding(t *testing.T) {

	test.Test(func() test.StorageManager {
		mgr := NewStorageManager("test")
		mgr.SetStateEncoding(record.StateEncodingFull)
		return mgr
	}, "filesystem", "", t)

}
