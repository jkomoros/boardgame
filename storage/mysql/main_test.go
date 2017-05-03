package mysql

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/test"
	"github.com/workfit/tester/assert"
	"log"
	"os"
	"testing"
)

//If outputTables is true, then will print create_tables.sql
const outputTables = false

func TestOutputTables(t *testing.T) {
	if !outputTables {
		return
	}

	filename := "create_tables.sql"

	log.Println("Outputing tables to", filename)

	if _, err := os.Stat(filename); err == nil {
		log.Println("That file already exists. Quitting. Delete it if you want to create a new one.")
		return
	}

	f, err := os.Create(filename)

	if err != nil {
		log.Println("Couldn't open file:", err)
	}

	defer f.Close()

	manager := NewStorageManager(true)
	manager.Connect("root:root@tcp(localhost:3306)/boardgame_test")

	logger := log.New(f, "", 0x0)

	manager.dbMap.TraceOn("", logger)

	manager.dbMap.CreateTablesIfNotExists()

}

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager(true)
	}, "mysql", "root:root@tcp(localhost:3306)/boardgame_test", t)

}

func TestWinnersConversion(t *testing.T) {
	tests := []struct {
		input       string
		result      []boardgame.PlayerIndex
		expectError bool
	}{
		{
			"",
			nil,
			false,
		},
		{
			"1,2",
			[]boardgame.PlayerIndex{1, 2},
			false,
		},
		{
			"-1",
			[]boardgame.PlayerIndex{-1},
			false,
		},
		{
			"1,2,",
			nil,
			true,
		},
	}

	for i, test := range tests {
		winners, err := stringToWinners(test.input)

		if test.expectError {
			assert.For(t, i).ThatActual(err).IsNotNil()
			continue
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
		}

		assert.For(t, i).ThatActual(winners).Equals(test.result).ThenDiffOnFail()

		reInput := winnersToString(test.result)

		assert.For(t, i).ThatActual(reInput).Equals(test.input)
	}
}
