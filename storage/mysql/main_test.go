package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/test"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	_ "github.com/mattes/migrate/source/file"
	"github.com/workfit/tester/assert"
	"log"
	"os"
	"testing"
)

const (
	testDSN          = "root:root@tcp(localhost:3306)/boardgame_test"
	pathToMigrations = "$GOPATH/src/github.com/jkomoros/boardgame/storage/mysql/migrations/"
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

	manager := NewStorageManager()
	manager.Connect("root:root@tcp(localhost:3306)/boardgame_test")

	logger := log.New(f, "", 0x0)

	manager.dbMap.TraceOn("", logger)

	manager.dbMap.CreateTablesIfNotExists()

}

func GetTestDatabase(t *testing.T) (*StorageManager, *migrate.Migrate) {

	dsn, err := getDSN(testDSN)

	if err != nil {
		t.Fatal(err)
		return nil, nil
	}

	path := os.ExpandEnv(pathToMigrations)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("The migrations path does not appear to exist")
		return nil, nil
	}

	db, _ := sql.Open("mysql", dsn)
	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path,
		"mysql",
		driver,
	)

	if err := m.Up(); err != nil {
		t.Fatal("Couldn't upgrade test database: ", err.Error())
		return nil, nil
	}

	return NewStorageManager(), m

}

func TestStorageManager(t *testing.T) {

	manager, m := GetTestDatabase(t)

	if manager == nil {
		//GetTestDatabase will have already fatal'd for us
		return
	}

	test.Test(func() test.StorageManager {
		return manager
	}, "mysql", testDSN, t)

	if err := m.Drop(); err != nil {
		log.Println(err)
	}

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
