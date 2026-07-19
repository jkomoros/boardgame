package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

type boardgameUtil struct {
	baseSubCommand
	Help            help
	Db              db
	Codegen         codegen
	Build           build
	Clean           clean
	Serve           serve
	Config          configCmd
	Stub            stubCmd
	Golden          goldenCmd
	EmitMoveNames   emitMoveNames
	EmitMoveArgs    emitMoveArgs
	EmitBoardSpaces emitBoardSpaces
	EmitTypes       emitTypes
	CheckClient     checkClient
	Lint            lintCmd
	Imagegen        imagegenCmd

	ConfigPath            string
	OverrideStarterConfig string

	config *config.Config

	//Dirs to delete on exit
	tempDirs []string

	cleanupMutex sync.Mutex
	cleanupOnce  sync.Once
	cleaning     bool
	processes    []*trackedProcess
}

type trackedProcess struct {
	process *os.Process
	done    chan struct{}
}

func (b *boardgameUtil) Run(p writ.Path, positional []string) {
	exitHelp(p.Last(), errors.New("COMMAND is required"))
}

func (b *boardgameUtil) Name() string {
	return "boardgame-util"
}

func (b *boardgameUtil) HelpText() string {

	return b.Name() +
		` is a comprehensive CLI tool to make working with
the boardgame framework easy. It has a number of subcommands to help do
everything from generate PropReader interfaces, to building and running a
server.

All of the substantive functionality provided by this utility is also
available as individual utility libraries to use directly if for some reason
this tool doesn't do exactly what you need.

A number of the commands expect some values to be provided in config.json. See
the README for more on configuring that configuration file, or run "boardgame-
util help config" to learn more.

See the individual sub-commands for more on what each one does.`

}

func (b *boardgameUtil) Usage() string {
	return "COMMAND [OPTION]... [ARG]..."
}

func (b *boardgameUtil) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"config", "c"},
			Decoder:     writ.NewOptionDecoder(&b.ConfigPath),
			Description: "The path to the config file or dir to use. If not provided, searches within current directory for files that could be a config, and then walks upwards until it finds one.",
		},
		{
			Names:       []string{"override-starter-config"},
			Decoder:     writ.NewOptionDecoder(&b.OverrideStarterConfig),
			Description: "If provided, the normal config will be ignored and a starter config will be used instead. Useful for running in contexts where you don't have a config.json set up yet. Valid values are the same as for `config init`",
			Placeholder: "TYPE",
		},
	}
}

func (b *boardgameUtil) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.Help,
		&b.Serve,
		&b.Config,
		&b.Codegen,
		&b.Stub,
		&b.Db,
		&b.Build,
		&b.Clean,
		&b.Golden,
		&b.EmitMoveNames,
		&b.EmitMoveArgs,
		&b.EmitBoardSpaces,
		&b.EmitTypes,
		&b.CheckClient,
		&b.Lint,
		&b.Imagegen,
	}
}

// Do any cleanup tasks as program exits.
func (b *boardgameUtil) Cleanup() {
	b.cleanupOnce.Do(func() {
		b.cleanupMutex.Lock()
		b.cleaning = true
		processes := append([]*trackedProcess(nil), b.processes...)
		dirs := append([]string(nil), b.tempDirs...)
		b.cleanupMutex.Unlock()

		for _, child := range processes {
			_ = terminateChildProcess(child.process)
		}

		allDone := make(chan struct{})
		go func() {
			for _, child := range processes {
				<-child.done
			}
			close(allDone)
		}()
		select {
		case <-allDone:
		case <-time.After(2 * time.Second):
			for _, child := range processes {
				_ = killChildProcess(child.process)
			}
			select {
			case <-allDone:
			case <-time.After(time.Second):
			}
		}

		for _, dir := range dirs {
			_ = os.RemoveAll(dir)
		}
	})
}

func (b *boardgameUtil) startTrackedProcess(cmd *exec.Cmd) (func(), error) {
	b.cleanupMutex.Lock()
	defer b.cleanupMutex.Unlock()
	if b.cleaning {
		return nil, errors.New("cannot start child process during cleanup")
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	child := &trackedProcess{process: cmd.Process, done: make(chan struct{})}
	b.processes = append(b.processes, child)
	var once sync.Once
	return func() { once.Do(func() { close(child.done) }) }, nil
}

func (b *boardgameUtil) errAndQuit(message string) {
	fmt.Fprintln(commandStderr, message)
	quit(1)
}

func (b *boardgameUtil) msgAndQuit(message string) {
	fmt.Fprintln(commandStdout, message)
	quit(0)
}

// NewTempDir vends an OS-managed temporary directory that is removed when the
// program exits. Keeping generated workspaces outside the caller's working
// directory prevents them from polluting (or accidentally entering) a repo.
func (b *boardgameUtil) NewTempDir(prefix string) string {
	dir, err := b.newTrackedTempDir(prefix)

	if err != nil {
		b.errAndQuit("Couldn't create temporary directory: " + err.Error())
	}

	return dir
}

func (b *boardgameUtil) newTrackedTempDir(prefix string) (string, error) {
	dir, err := newSystemTempDir(prefix)
	if err != nil {
		return "", err
	}
	b.cleanupMutex.Lock()
	defer b.cleanupMutex.Unlock()
	if b.cleaning {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			return "", fmt.Errorf("cleanup already started; remove untracked temp directory: %w", removeErr)
		}
		return "", errors.New("cleanup already started")
	}
	b.tempDirs = append(b.tempDirs, dir)
	return dir, nil
}

func newSystemTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

func (b *boardgameUtil) starterConfigForType(typ string) (*config.Config, error) {

	if typ == "" {
		typ = "default"
	}

	typ = strings.ToLower(typ)

	configPath := b.ConfigPath

	switch typ {
	case "default":
		return config.DefaultStarterConfig(configPath), nil
	case "sample":
		return config.SampleStarterConfig(configPath), nil
	case "minimal":
		return config.MinimalStarterConfig(configPath), nil
	default:
		return nil, errors.New(typ + " is not a legal type")
	}
}

// GetConfig fetches the config, finding it from disk if it hasn't yet. If
// finding the config errors for any reason, program will quit. That is, when
// you call this method we assume that it's required for operation of that
// command.
func (b *boardgameUtil) GetConfig(createIfNotExist bool) *config.Config {
	if b.config != nil {
		return b.config
	}

	var c *config.Config
	var err error

	if b.OverrideStarterConfig != "" {
		fmt.Println("Ignoring normal config, using starter config of type: " + b.OverrideStarterConfig)
		c, err = b.starterConfigForType(b.OverrideStarterConfig)
		if err != nil {
			b.errAndQuit(err.Error())
			return nil
		}
	} else {
		c, err = config.Get(b.ConfigPath, createIfNotExist)
	}

	if err != nil {
		b.errAndQuit("config is required for this command, but it couldn't be loaded. You can create one with `boardgame-util config init`.\nError: " + err.Error())
	}

	b.config = c

	return c
}
