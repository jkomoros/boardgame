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

// if pkgs == nil, will use the game packages from the selected mode.
func (s *serve) doServe(p writ.Path, positional []string, pkgs []*gamepkg.Pkg, storageLiteralArgs string) {

	c := s.Base().GetConfig(false)

	if s.OfflineDevMode {
		c.AddOverride(config.EnableOfflineDevMode())
	}

	mode := c.Dev

	if s.Prod {
		mode = c.Prod
	}

	staticPort := mode.DefaultStaticPort
	if s.StaticPort != "" {
		staticPort = s.StaticPort
	}

	port := mode.DefaultPort
	if s.Port != "" {
		port = s.Port
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
		StorageLiteralArgs:     storageLiteralArgs,
		OverrideAllowedOrigins: localServeAllowedOrigins(staticPort),
	}

	if s.OfflineDevMode {
		apiOptions.OverrideOfflineDevMode = true
	}

	fmt.Println("Generating move name constants")
	if err := emitMoveNamesForPackages(s.Base(), pkgs); err != nil {
		s.Base().errAndQuit("Couldn't generate move names: " + err.Error())
	}

	fmt.Println("Generating move argument types")
	if err := emitMoveArgsForPackages(s.Base(), pkgs, false); err != nil {
		s.Base().errAndQuit("Couldn't generate move args: " + err.Error())
	}
	if err := emitBoardSpacesForPackages(pkgs, false); err != nil {
		s.Base().errAndQuit("Couldn't emit authored board spaces: " + err.Error())
	}

	fmt.Println("Generating type definitions")
	if err := emitTypesForPackages(s.Base(), pkgs); err != nil {
		s.Base().errAndQuit("Couldn't generate type definitions: " + err.Error())
	}

	fmt.Println("Creating temporary binary")
	apiPath, err := api.Build(dir, pkgs, storage, apiOptions)

	if err != nil {
		s.Base().errAndQuit("Couldn't create api: " + err.Error())
	}

	fmt.Println("Creating temporary static assets folder")
	//TODO: should we allow you to pass CopyFiles? I don't know why you'd want
	//to given this is a temp dir.
	clientConfig := clientConfigForServe(c, s.Prod)
	_, err = static.Build(dir, pkgs, clientConfig, s.Prod, false, mode.OfflineDevMode)

	if err != nil {
		s.Base().errAndQuit("Couldn't create static directory: " + err.Error())
	}

	type childResult struct {
		name string
		err  error
	}
	results := make(chan childResult, 2)

	// Start both children before waiting on either. Whichever child exits first
	// ends the whole serve session; Cleanup also kills both on SIGINT/SIGTERM.
	fmt.Println("Starting up Vite dev server on port " + staticPort)
	viteCmd := static.ServerCommand(dir, staticPort, port)
	viteDone, err := s.Base().startTrackedProcess(viteCmd)
	if err != nil {
		s.Base().errAndQuit("Couldn't start Vite dev server: " + err.Error())
	}
	go func() {
		err := viteCmd.Wait()
		viteDone()
		results <- childResult{"Vite", err}
	}()

	//cmd will be run as though it's in this directory, which is where
	//config.json is.
	cmd := exec.Command(apiPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "PORT="+port)
	configureChildProcess(cmd)

	apiDone, err := s.Base().startTrackedProcess(cmd)
	if err != nil {
		_ = terminateChildProcess(viteCmd.Process)
		select {
		case <-results:
		case <-time.After(2 * time.Second):
			_ = killChildProcess(viteCmd.Process)
			<-results
		}
		s.Base().errAndQuit("Couldn't start API server: " + err.Error())
	}

	go func() {
		err := cmd.Wait()
		apiDone()
		results <- childResult{"API", err}
	}()

	topLine := "************************************************************************"

	// Preserve the short grace period before advertising the URL. Automated
	// callers use an API-backed readiness probe instead of relying on this text.
	time.Sleep(time.Second * 2)

	fmt.Println(" ")
	fmt.Println(topLine)
	for i := 0; i < 2; i++ {
		fmt.Println("*")
	}
	fmt.Println("*     Server running. Open 'http://127.0.0.1:" + staticPort + "' in your browser")
	for i := 0; i < 2; i++ {
		fmt.Println("*")
	}
	fmt.Println(topLine)
	fmt.Println(" ")

	first := <-results
	var sibling *os.Process
	if first.name == "Vite" {
		sibling = cmd.Process
	} else {
		sibling = viteCmd.Process
	}
	_ = terminateChildProcess(sibling)
	// Reap the sibling before deleting its temporary working directory. Escalate
	// only if the process group ignores a graceful termination request.
	select {
	case <-results:
	case <-time.After(2 * time.Second):
		_ = killChildProcess(sibling)
		<-results
	}
	if first.err != nil {
		s.Base().errAndQuit(first.name + " server exited: " + first.err.Error())
	}
	s.Base().errAndQuit(first.name + " server exited unexpectedly")
}

func localServeAllowedOrigins(staticPort string) string {
	return "http://localhost:" + staticPort + ",http://127.0.0.1:" + staticPort
}

// clientConfigForServe returns the generated browser configuration for a serve
// invocation. A local serve session always talks to its API through Vite's
// same-origin proxy. Both host variants must be empty:
// the legacy bootstrap selects between them by hostname, and an isolated serve
// may bind 127.0.0.1 rather than the literal hostname "localhost". This makes
// arbitrary test/dev ports work without weakening the API's CORS allowlist.
func clientConfigForServe(c *config.Config, prodMode bool) *config.ClientConfig {
	result := c.Client(prodMode)
	if result == nil {
		return result
	}
	clone := *result
	clone.Host = ""
	clone.DevHost = ""
	return &clone
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
