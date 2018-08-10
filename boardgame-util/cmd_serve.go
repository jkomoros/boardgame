package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"io/ioutil"
	"os/exec"
)

type Serve struct {
	baseSubCommand
}

func (s *Serve) Run(p writ.Path, positional []string) {

	base := s.Base().(*BoardgameUtil)

	config := base.GetConfig()
	mode := config.Dev

	//TODO: this should be in a Base helper to get a new temp directory that
	//base will take care of kiling.
	dir, err := ioutil.TempDir(".", "serve_")

	if err != nil {
		errAndQuit("Couldn't create temporary directory: " + err.Error())
	}

	//TODO: remove dir on exit
	//TODO: allow specifying a different storage type

	apiPath, err := build.Api(dir, mode.GamesList, build.StorageBolt)

	if err != nil {
		errAndQuit("Couldn't create api: " + err.Error())
	}

	//cmd will be run as though it's in this directory, which is where
	//config.json is.
	cmd := exec.Command(apiPath)
	//TODO: connect StdOut and StdErr to our own

	err = cmd.Run()

	errAndQuit("Error running command: " + err.Error())
}

func (s *Serve) Name() string {
	return "serve"
}

func (s *Serve) Description() string {
	return "Creates and runs a local development server based on config.json"
}
