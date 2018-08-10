package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"os"
	"os/exec"
)

type Serve struct {
	baseSubCommand
}

func (s *Serve) Run(p writ.Path, positional []string) {

	base := s.Base().(*BoardgameUtil)

	config := base.GetConfig()
	mode := config.Dev

	dir := base.NewTempDir("temp_serve_")

	//TODO: allow specifying a different storage type

	apiPath, err := build.Api(dir, mode.GamesList, build.StorageBolt)

	if err != nil {
		errAndQuit("Couldn't create api: " + err.Error())
	}

	//cmd will be run as though it's in this directory, which is where
	//config.json is.
	cmd := exec.Command(apiPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()

	errAndQuit("Error running command: " + err.Error())
}

func (s *Serve) Name() string {
	return "serve"
}

func (s *Serve) Description() string {
	return "Creates and runs a local development server based on config.json"
}
