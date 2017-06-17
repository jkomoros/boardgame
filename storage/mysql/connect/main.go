package connect

import (
	"database/sql"
	"errors"
	dsnparser "github.com/go-sql-driver/mysql"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	_ "github.com/mattes/migrate/source/file"
	"os"
)

const (
	pathToMigrations = "$GOPATH/src/github.com/jkomoros/boardgame/storage/mysql/migrations/"
	testDbName       = "TEMPORARY_DATABASE_boardgame_test"
)

func Db(dsn string, testMode bool, createDb bool) (*sql.DB, error) {

	dbName := "boardgame"

	if testMode {
		dbName = testDbName
	}

	//TODO: if createDb is true, make sure the DB exists and create if not.

	parsedDSN, err := dsnparser.ParseDSN(dsn)

	if err != nil {
		return nil, errors.New("config provided was not valid DSN: " + err.Error())
	}

	parsedDSN.Collation = "utf8mb4_unicode_ci"
	parsedDSN.MultiStatements = true

	if createDb {
		oldDbName := parsedDSN.DBName
		parsedDSN.DBName = ""
		if err := doCreateDb(parsedDSN.FormatDSN(), dbName); err != nil {
			return nil, errors.New("Couldn't create database: " + err.Error())
		}
		parsedDSN.DBName = oldDbName
	}

	if parsedDSN.DBName != dbName {
		return nil, errors.New("DBName did not mach expectations. Got " + parsedDSN.DBName + " expected " + dbName)
	}

	dsnToUse := parsedDSN.FormatDSN()

	db, err := sql.Open("mysql", dsnToUse)

	if err != nil {
		return nil, errors.New("Unable to open database: " + err.Error())
	}

	return db, nil
}

func doCreateDb(dsn string, dbName string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return errors.New("couldn't open database: " + err.Error())
	}

	defer db.Close()

	_, err = db.Exec("create database if not exists `" + dbName + "`;")
	if err != nil {
		return errors.New("Couldn't create database: " + err.Error())
	}
	return nil
}

func Migrations(db *sql.DB) (*migrate.Migrate, error) {
	path := os.ExpandEnv(pathToMigrations)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("The migrations folder does not appear to exist: " + err.Error())
	}

	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path,
		"mysql",
		driver,
	)

	if err != nil {
		return nil, errors.New("Couldnt' create migration instance: " + err.Error())
	}
	return m, nil
}

func DropTestDb(dsn string) error {

	parsedDSN, err := dsnparser.ParseDSN(dsn)

	if err != nil {
		return errors.New("config provided was not valid DSN: " + err.Error())
	}

	if parsedDSN.DBName != testDbName {
		return errors.New("Database to connect to wasn't test db name")
	}

	db, err := sql.Open("mysql", parsedDSN.FormatDSN())

	if err != nil {
		return errors.New("Couldn't connect to db: " + err.Error())
	}

	defer db.Close()

	_, err = db.Exec("drop database `" + testDbName + "`;")

	if err != nil {
		return errors.New("Couldnt' drop test db: " + err.Error())
	}
	return nil
}
