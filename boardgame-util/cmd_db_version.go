package main

import (
	"fmt"
	"strconv"

	"github.com/bobziuchkovski/writ"
)

type dbVersion struct {
	baseSubCommand
}

func (d *dbVersion) Name() string {
	return "version"
}

func (d *dbVersion) Description() string {
	return "Print what version of the migrations have been applied to db so far and quit"
}

func (d *dbVersion) Run(p writ.Path, positonal []string) {
	parent := d.Parent().(*Db)

	m := parent.GetMigrate(false)

	version, _, err := m.Version()

	if err != nil {
		d.Base().errAndQuit("Couldn't get version: " + err.Error())
	}

	fmt.Println("Version: " + strconv.Itoa(int(version)))

}
