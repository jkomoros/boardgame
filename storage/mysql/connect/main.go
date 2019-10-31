package connect

import (
	"database/sql"
	"errors"
	"os"

	dsnparser "github.com/go-sql-driver/mysql"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"

	//This is the way to include the necessary driver
	_ "github.com/mattes/migrate/source/file"
)

const (
	pathToMigrations = "$GOPATH/src/github.com/jkomoros/boardgame/storage/mysql/migrations/"
	//TestDbName is the name of the test database to create.
	TestDbName = "TEMPORARY_DATABASE_boardgame_test"
)

//Db creates the Database handle. If testMode is true, then dbName will be
//TestDbName.
func Db(dsn string, testMode bool, createDb bool) (*sql.DB, error) {

	dbName := "boardgame"

	if testMode {
		dbName = TestDbName
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

//Migrations returns the Migrate object to migrate a database.
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

//DropTestDb is called to delete the test database. Will fail if the database name isn't TestDbName
func DropTestDb(dsn string) error {

	parsedDSN, err := dsnparser.ParseDSN(dsn)

	if err != nil {
		return errors.New("config provided was not valid DSN: " + err.Error())
	}

	if parsedDSN.DBName != TestDbName {
		return errors.New("Database to connect to wasn't test db name")
	}

	db, err := sql.Open("mysql", parsedDSN.FormatDSN())

	if err != nil {
		return errors.New("Couldn't connect to db: " + err.Error())
	}

	defer db.Close()

	_, err = db.Exec("drop database `" + TestDbName + "`;")

	if err != nil {
		return errors.New("Couldnt' drop test db: " + err.Error())
	}
	return nil
}
