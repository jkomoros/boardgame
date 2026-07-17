package mysql

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/storage/internal/test"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
	"github.com/mattes/migrate"
	"github.com/workfit/tester/assert"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jkomoros/boardgame/server/api/tablelease"
)

func TestRematchMetadataConversionRoundTrip(t *testing.T) {
	want := &extendedgame.StorageRecord{
		Open: true, Visible: true, Owner: "owner", CompanionRoomCode: "PLAY",
		CompanionLocked: true, RematchGameID: "0123456789abcdef", RematchReady: true,
	}
	got := newExtendedGameStorageRecord(want).ToStorageRecord()
	if *got != *want {
		t.Fatalf("extended rematch round trip = %+v; want %+v", got, want)
	}

	combined := &extendedgame.CombinedStorageRecord{StorageRecord: *want}
	combinedGot := newCombinedGameStorageRecord(combined).ToStorageRecord()
	if combinedGot.RematchGameID != want.RematchGameID || !combinedGot.RematchReady {
		t.Fatalf("combined rematch round trip = %+v", combinedGot.StorageRecord)
	}
}

const (
	testDSN          = "root:root@tcp(localhost:3306)/TEMPORARY_DATABASE_boardgame_test"
	pathToMigrations = "$GOPATH/src/github.com/jkomoros/boardgame/storage/mysql/migrations/"
)

func TestTableLeaseTransferSchemaCompatibilityBoundary(t *testing.T) {
	if tableLeaseUsesTransferColumns(&tablelease.StorageRecord{}) {
		t.Fatal("ordinary lease unexpectedly requires migration 0023")
	}
	for _, record := range []*tablelease.StorageRecord{
		{TransferID: "x"}, {PreviousDeviceID: "x"}, {TransitionKind: tablelease.TransitionHostAction},
	} {
		if !tableLeaseUsesTransferColumns(record) {
			t.Fatalf("transfer-capable record did not require migration 0023: %+v", record)
		}
	}
}

func TestTableLeaseTransferMigrationRoundTrip(t *testing.T) {
	if _, err := net.DialTimeout("tcp", "localhost:3306", 200*time.Millisecond); err != nil {
		t.Skip("MySQL not available at localhost:3306")
	}
	db, err := connect.Db(testDSN, true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrations, err := connect.Migrations(db)
	if err != nil {
		t.Fatal(err)
	}
	defer migrations.Close()
	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}
	manager := &StorageManager{db: db}
	defer func() {
		if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
			t.Errorf("restore latest migration: %v", err)
		}
	}()
	if err := migrations.Steps(-1); err != nil {
		t.Fatal(err)
	}
	available, err := manager.companionTableTransferColumnsAvailable()
	if err != nil || available {
		t.Fatalf("0023 down availability/error = %t/%v", available, err)
	}
	if err := migrations.Steps(1); err != nil {
		t.Fatal(err)
	}
	available, err = manager.companionTableTransferColumnsAvailable()
	if err != nil || !available {
		t.Fatalf("0023 up availability/error = %t/%v", available, err)
	}
}

// If outputTables is true, then will print create_tables.sql
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

func GetTestDatabase(t *testing.T) *StorageManager {

	db, err := connect.Db(testDSN, true, true)

	if err != nil {
		t.Fatal("Couldn't get db: " + err.Error())
		return nil
	}

	m, err := connect.Migrations(db)

	if err != nil {
		t.Fatal("Couldn't get migrations: " + err.Error())
		return nil
	}

	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			t.Fatal("Couldn't upgrade test database: ", err.Error())
			return nil
		}
	}

	m.Close()

	return NewStorageManager(true)

}

func TestStorageManager(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:3306", 250*time.Millisecond)
	if err != nil {
		t.Skip("MySQL integration test requires a server on localhost:3306")
	}
	conn.Close()

	test.Test(func() test.StorageManager {
		return GetTestDatabase(t)
	}, "mysql", testDSN, t)

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
