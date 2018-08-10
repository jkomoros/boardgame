package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"strconv"
)

type DbVersion struct {
	baseSubCommand
}

func (d *DbVersion) Name() string {
	return "version"
}

func (d *DbVersion) Description() string {
	return "Print what version of the migrations have been applied to db so far and quit"
}

func (d *DbVersion) Run(p writ.Path, positonal []string) {
	parent := d.Parent().(*Db)

	m := parent.GetMigrate(false)

	version, _, err := m.Version()

	if err != nil {
		errAndQuit("Couldn't get version: " + err.Error())
	}

	fmt.Println("Version: " + strconv.Itoa(int(version)))

}
