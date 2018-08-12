package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Serve struct {
	baseSubCommand

	Storage string
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
	staticPath, err := build.Static(dir, mode.GamesList, config)

	if err != nil {
		errAndQuit("Couldn't create static directory: " + err.Error())
	}

	go func() {

		fs := http.FileServer(http.Dir(staticPath))

		staticPort := "8080"

		if mode.DefaultStaticPort != "" {
			staticPort = mode.DefaultStaticPort
		}

		http.Handle("/", fs)
		fmt.Println("Starting up asset server at " + staticPort)
		http.ListenAndServe(":"+staticPort, nil)
	}()

	//TODO: simple serving of staticPath here. Do we need a new parameter for
	//default static serving port?

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

func (s *Serve) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"storage", "s"},
			Decoder:     writ.NewOptionDecoder(&s.Storage),
			Description: "Which storage subsystem to use. One of {" + strings.Join(build.ValidStorageTypeStrings(), ",") + "}. If not provided, falls back on the DefaultStorageType from config, or as a final fallback just the deafult storage type.",
		},
	}
}
