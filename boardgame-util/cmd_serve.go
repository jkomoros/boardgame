package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Serve struct {
	baseSubCommand

	Storage string

	Port       string
	StaticPort string
	ForceBower bool
	Prod       bool
}

func (s *Serve) Run(p writ.Path, positional []string) {

	config := s.Base().GetConfig(false)
	mode := config.Dev

	dir := s.Base().NewTempDir("temp_serve_")

	storage := effectiveStorageType(s.Base(), mode, s.Storage)

	fmt.Println("Creating temporary binary")
	apiPath, err := build.Api(dir, mode.Games, storage)

	if err != nil {
		s.Base().errAndQuit("Couldn't create api: " + err.Error())
	}

	fmt.Println("Creating temporary static assets folder")
	//TODO: should we allow you to pass CopyFiles? I don't know why you'd want
	//to given this is a temp dir.
	_, err = build.Static(dir, mode.Games, config, s.ForceBower, s.Prod, false)

	if err != nil {
		s.Base().errAndQuit("Couldn't create static directory: " + err.Error())
	}

	staticPort := mode.DefaultStaticPort

	if s.StaticPort != "" {
		staticPort = s.StaticPort
	}

	go func() {
		fmt.Println("Starting up asset server at " + staticPort)
		if err := build.StaticServer(dir, staticPort); err != nil {
			//TODO: when this happens we should quit the whole program
			fmt.Println("ERROR: couldn't start static server: " + err.Error())
		}
	}()

	//TODO: simple serving of staticPath here. Do we need a new parameter for
	//default static serving port?

	port := mode.DefaultPort

	if s.Port != "" {
		port = s.Port
	}

	//cmd will be run as though it's in this directory, which is where
	//config.json is.
	cmd := exec.Command(apiPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "PORT="+port)

	err = cmd.Start()

	if err == nil {
		topLine := "************************************************************************"

		//Cheat and wait to print the message until later
		time.Sleep(time.Second * 2)

		fmt.Println(" ")
		fmt.Println(topLine)
		for i := 0; i < 2; i++ {
			fmt.Println("*")
		}
		fmt.Println("*     Server running. Open 'http://localhost:" + staticPort + "' in your browser")
		for i := 0; i < 2; i++ {
			fmt.Println("*")
		}
		fmt.Println(topLine)
		fmt.Println(" ")

		err = cmd.Wait()
	}

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			s.Base().errAndQuit("Couldn't cast exiterror")
		}

		//Programs that are signaled and who responded to it before us (the
		//parent) did (which is a race) will have Exited() false, whereas a
		//program that errored and quit on its own should have true. Only the
		//latter is an err; calling errAndQuit not in an error could prevent
		//our own clean shutdown from happening.
		if exitErr.ProcessState.Exited() {
			s.Base().errAndQuit("Error running command: " + err.Error())
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
		{
			Names:       []string{"force-bower"},
			Description: "If provided, will force an update to bower_components even if that folder already exists.",
			Decoder:     writ.NewFlagDecoder(&s.ForceBower),
			Flag:        true,
		},
		{
			Names:       []string{"prod"},
			Description: "If provided, will created bundled build directory for static resources.",
			Decoder:     writ.NewFlagDecoder(&s.Prod),
			Flag:        true,
		},
	}
}
