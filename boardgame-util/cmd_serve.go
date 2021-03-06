package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/api"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type serve struct {
	baseSubCommand

	Storage string

	Port       string
	StaticPort string
	Prod       bool

	OfflineDevMode bool
}

//if pkgs == nil, will use the game packages from the selected mode.
func (s *serve) doServe(p writ.Path, positional []string, pkgs []*gamepkg.Pkg, storageLiteralArgs string) {

	c := s.Base().GetConfig(false)

	if s.OfflineDevMode {
		c.AddOverride(config.EnableOfflineDevMode())
	}

	mode := c.Dev

	if s.Prod {
		mode = c.Prod
	}

	dir := s.Base().NewTempDir("temp_serve_")

	storage := effectiveStorageType(s.Base(), mode, s.Storage)

	if pkgs == nil {
		var err error
		pkgs, err = mode.AllGamePackages()

		if err != nil {
			s.Base().errAndQuit("Not all game packages were valid: " + err.Error())
		}
	}

	apiOptions := &api.Options{
		StorageLiteralArgs: storageLiteralArgs,
	}

	if s.OfflineDevMode {
		apiOptions.OverrideOfflineDevMode = true
	}

	fmt.Println("Creating temporary binary")
	apiPath, err := api.Build(dir, pkgs, storage, apiOptions)

	if err != nil {
		s.Base().errAndQuit("Couldn't create api: " + err.Error())
	}

	fmt.Println("Creating temporary static assets folder")
	//TODO: should we allow you to pass CopyFiles? I don't know why you'd want
	//to given this is a temp dir.
	_, err = static.Build(dir, pkgs, c.Client(s.Prod), s.Prod, false, mode.OfflineDevMode)

	if err != nil {
		s.Base().errAndQuit("Couldn't create static directory: " + err.Error())
	}

	staticPort := mode.DefaultStaticPort

	if s.StaticPort != "" {
		staticPort = s.StaticPort
	}

	go func() {
		fmt.Println("Starting up asset server at " + staticPort)
		if err := static.Server(dir, staticPort); err != nil {
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

func (s *serve) Run(p writ.Path, positional []string) {
	s.doServe(p, positional, nil, "")
}

func (s *serve) Name() string {
	return "serve"
}

func (s *serve) Description() string {
	return "Creates and runs a local development server based on config.json"
}

func (s *serve) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"storage", "s"},
			Decoder:     writ.NewOptionDecoder(&s.Storage),
			Description: "Which storage subsystem to use. One of {" + strings.Join(api.ValidStorageTypeStrings(), ",") + "}. If not provided, falls back on the DefaultStorageType from config, or as a final fallback just the deafult storage type.",
		},
		{
			Names:       []string{"prod"},
			Description: "If provided, will created bundled build directory for static resources.",
			Decoder:     writ.NewFlagDecoder(&s.Prod),
			Flag:        true,
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
			Names:       []string{"offline-dev-mode"},
			Description: "If provided, will override OfflineDevMode to true, no matter what is in config. The effect of this is that the webapp won't make any calls to anything but localhost, allowing development on for example a plane. This is generally the best way to enable offline dev mode.",
			Decoder:     writ.NewFlagDecoder(&s.OfflineDevMode),
			Flag:        true,
		},
	}
}
