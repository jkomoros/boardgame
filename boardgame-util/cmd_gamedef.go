package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

// gamedefDir is where gamedef export writes, relative to the repository root.
const gamedefDir = "conformance/definitions"

// gamedefGame is one game gamedef export knows how to build. The run-time
// half needs a live delegate, so only games compiled into boardgame-util can
// be exported.
type gamedefGame struct {
	name        string
	importPath  string
	newDelegate func() boardgame.GameDelegate
}

var gamedefGames = []gamedefGame{
	{"memory", "github.com/jkomoros/boardgame/examples/memory", func() boardgame.GameDelegate { return memory.NewDelegate() }},
	{"blackjack", "github.com/jkomoros/boardgame/examples/blackjack", func() boardgame.GameDelegate { return blackjack.NewDelegate() }},
	{"pig", "github.com/jkomoros/boardgame/examples/pig", func() boardgame.GameDelegate { return pig.NewDelegate() }},
	{"tictactoe", "github.com/jkomoros/boardgame/examples/tictactoe", func() boardgame.GameDelegate { return tictactoe.NewDelegate() }},
	{"werewolf", "github.com/jkomoros/boardgame/examples/werewolf", func() boardgame.GameDelegate { return werewolf.NewDelegate() }},
	{"debuganimations", "github.com/jkomoros/boardgame/examples/debuganimations", func() boardgame.GameDelegate { return debuganimations.NewDelegate() }},
	{"checkers", "github.com/jkomoros/boardgame/examples/checkers", func() boardgame.GameDelegate { return checkers.NewDelegate() }},
}

type gamedefCmd struct {
	baseSubCommand

	Export gamedefExport
}

func (c *gamedefCmd) Run(p writ.Path, positional []string) {
	exitHelp(p.Last(), errors.New("SUBCOMMAND is required"))
}

func (c *gamedefCmd) Name() string {
	return "gamedef"
}

func (c *gamedefCmd) Description() string {
	return "Exports each example game's definition as data for the TypeScript engine"
}

func (c *gamedefCmd) HelpText() string {
	return c.Name() + ` works with exported game definitions: one JSON document per game (GameData, boardgame-util/lib/gamedef) holding everything about the game that is already data. See fabric/TRANSPILE-PROPOSAL.md, "The exported definition".`
}

func (c *gamedefCmd) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.Export,
	}
}

type gamedefExport struct {
	baseSubCommand

	Check bool
	Root  string
}

// exportGamedef builds one game's GameData and its checked-in bytes. A
// static half that is not implemented yet is tolerated: note says so, and
// the run-time parts are exported alone.
func exportGamedef(game gamedefGame) (blob []byte, note string, err error) {
	manager, err := boardgame.NewGameManager(game.newDelegate(), memstore.NewStorageManager())
	if err != nil {
		return nil, "", fmt.Errorf("create manager: %w", err)
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	static, err := gamedef.Static(game.importPath)
	if err != nil {
		if !strings.Contains(err.Error(), "not implemented") {
			return nil, "", fmt.Errorf("static half: %w", err)
		}
		note = game.name + ": " + err.Error() + "; exporting the run-time parts only"
		static = nil
	}
	data, err := gamedef.Export(manager, static)
	if err != nil {
		return nil, note, err
	}
	blob, err = gamedef.Marshal(data)
	return blob, note, err
}

func (c *gamedefExport) Run(p writ.Path, positional []string) {
	root := c.Root
	if root == "" {
		found, err := boardgameRepoRoot()
		if err != nil {
			c.Base().errAndQuit(err.Error())
		}
		root = found
	}

	games := gamedefGames
	all := len(positional) == 0
	if !all {
		games = nil
		for _, arg := range positional {
			var found *gamedefGame
			for i := range gamedefGames {
				if gamedefGames[i].name == arg || gamedefGames[i].importPath == arg {
					found = &gamedefGames[i]
				}
			}
			if found == nil {
				exitHelp(p.Last(), errors.New(arg+" is not a game gamedef export knows; it knows the seven example games, by name or import path"))
			}
			games = append(games, *found)
		}
	}

	dir := filepath.Join(root, filepath.FromSlash(gamedefDir))
	var problems []string
	written := 0
	expected := map[string]bool{}
	for _, game := range games {
		blob, note, err := exportGamedef(game)
		if note != "" {
			fmt.Fprintln(commandStderr, "Note: "+note)
		}
		if err != nil {
			c.Base().errAndQuit("Couldn't export " + game.name + ": " + err.Error())
		}
		file := game.name + ".json"
		expected[file] = true
		path := filepath.Join(dir, file)
		if c.Check {
			got, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				problems = append(problems, file+": missing")
				continue
			}
			if err != nil {
				c.Base().errAndQuit(err.Error())
			}
			if !bytes.Equal(got, blob) {
				problems = append(problems, file+": differs from what the exporter produces now")
			}
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.Base().errAndQuit(err.Error())
		}
		if err := os.WriteFile(path, blob, 0644); err != nil {
			c.Base().errAndQuit(err.Error())
		}
		written++
	}
	if c.Check && all {
		onDisk, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			c.Base().errAndQuit(err.Error())
		}
		for _, path := range onDisk {
			if !expected[filepath.Base(path)] {
				problems = append(problems, filepath.Base(path)+": checked in but not exported")
			}
		}
	}
	if c.Check {
		if len(problems) > 0 {
			sort.Strings(problems)
			for _, problem := range problems {
				fmt.Fprintln(commandStderr, problem)
			}
			c.Base().errAndQuit("The game definitions are stale; run `boardgame-util gamedef export`")
		}
		fmt.Println("The game definitions are current")
		return
	}
	fmt.Printf("Wrote %d files to %s\n", written, dir)
}

func (c *gamedefExport) Name() string {
	return "export"
}

func (c *gamedefExport) Description() string {
	return "Regenerates the checked-in game definitions"
}

func (c *gamedefExport) HelpText() string {
	return c.Name() + ` [PKG...] exports each named example game (by name, such as memory, or import path; all seven by default) to ` + gamedefDir + `/<game>.json at the repository root. It joins the run-time half (the live manager) with the static half (the game's source); while the static half is not implemented, the run-time parts are exported alone and a note is printed.

The output is deterministic. With --check nothing is written; the command fails if a checked-in definition differs from what the exporter produces now.`
}

func (c *gamedefExport) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"check"}, Description: "Verify the checked-in definitions are current without writing files.", Decoder: writ.NewFlagDecoder(&c.Check), Flag: true},
		{Names: []string{"root"}, Description: "The boardgame repository root. Defaults to the one containing the working directory.", Decoder: writ.NewOptionDecoder(&c.Root), Placeholder: "DIR"},
	}
}
