package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"os"
	"os/exec"
	"strings"
)

type Serve struct {
	baseSubCommand

	Storage string

	Port       string
	StaticPort string
}

func (s *Serve) Run(p writ.Path, positional []string) {

	base := s.Base().(*BoardgameUtil)

	config := base.GetConfig()
	mode := config.Dev

	dir := base.NewTempDir("temp_serve_")

	storage := effectiveStorageType(mode, s.Storage)

	fmt.Println("Creating temporary binary")
	apiPath, err := build.Api(dir, mode.GamesList, storage)

	if err != nil {
		errAndQuit("Couldn't create api: " + err.Error())
	}

	fmt.Println("Creating temporary static assets folder")
	_, err = build.Static(dir, mode.GamesList, config)

	if err != nil {
		errAndQuit("Couldn't create static directory: " + err.Error())
	}

	staticPort := "8080"

	if mode.DefaultStaticPort != "" {
		staticPort = mode.DefaultStaticPort
	}

	if s.StaticPort != "" {
		staticPort = s.StaticPort
	}

	go func() {
		fmt.Println("Starting up asset server at " + staticPort)
		if err := build.SimpleStaticServer(dir, staticPort); err != nil {
			//TODO: when this happens we should quit the whole program
			fmt.Println("ERROR: couldn't start static server: " + err.Error())
		}
	}()

	//TODO: simple serving of staticPath here. Do we need a new parameter for
	//default static serving port?

	port := "8888"

	if mode.DefaultPort != "" {
		port = mode.DefaultPort
	}

	if s.Port != "" {
		port = s.Port
	}

	//cmd will be run as though it's in this directory, which is where
	//config.json is.
	cmd := exec.Command(apiPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "PORT="+port)

	err = cmd.Run()

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			errAndQuit("Couldn't cast exiterror")
		}

		//Programs that are signaled and who responded to it before us (the
		//parent) did (which is a race) will have Exited() false, whereas a
		//program that errored and quit on its own should have true. Only the
		//latter is an err; calling errAndQuit not in an error could prevent
		//our own clean shutdown from happening.
		if exitErr.ProcessState.Exited() {
			errAndQuit("Error running command: " + err.Error())
		}
	}
}

func (s *Serve) Name() string {
	return "serve"
}

func (s *Serve) Description() string {
	return "Creates and runs a local development server based on config.json"
}

func (s *Serve) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"storage", "s"},
			Decoder:     writ.NewOptionDecoder(&s.Storage),
			Description: "Which storage subsystem to use. One of {" + strings.Join(build.ValidStorageTypeStrings(), ",") + "}. If not provided, falls back on the DefaultStorageType from config, or as a final fallback just the deafult storage type.",
		},
		{
			Names:       []string{"port", "p"},
			Decoder:     writ.NewOptionDecoder(&s.Port),
			Description: "Port to use for the api server, overriding value in config.json's DefaultPort",
		},
		{
			Names:       []string{"static-port"},
			Decoder:     writ.NewOptionDecoder(&s.StaticPort),
			Description: "Port to use for the static file server, overridig value in config.json's DefaultStaticPort",
		},
	}
}
