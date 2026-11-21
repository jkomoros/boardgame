package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance/corpus"
)

type conformanceCmd struct {
	baseSubCommand

	Export conformanceExport
}

func (c *conformanceCmd) Run(p writ.Path, positional []string) {
	exitHelp(p.Last(), errors.New("SUBCOMMAND is required"))
}

func (c *conformanceCmd) Name() string {
	return "conformance"
}

func (c *conformanceCmd) Description() string {
	return "Exports complete per-viewer records of example games for checking another engine implementation"
}

func (c *conformanceCmd) HelpText() string {
	return c.Name() + ` works with the conformance corpus: documents that record, for every state version of a game run on the real engine, the unsanitized state, the move, and what every viewer is sent. See boardgame-util/lib/conformance and conformance/schema.ts.`
}

func (c *conformanceCmd) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.Export,
	}
}

type conformanceExport struct {
	baseSubCommand

	Check bool
	Plain bool
	Root  string
	Out   string
}

func (c *conformanceExport) Run(p writ.Path, positional []string) {
	if len(positional) > 0 {
		exitHelp(p.Last(), errors.New(c.Name()+" doesn't take any positional parameters"))
	}
	root := c.Root
	if root == "" {
		found, err := boardgameRepoRoot()
		if err != nil {
			c.Base().errAndQuit(err.Error())
		}
		root = found
	}
	if c.Check {
		problems, err := corpus.Check(root)
		if err != nil {
			c.Base().errAndQuit("Couldn't generate the corpus: " + err.Error())
		}
		if len(problems) > 0 {
			for _, problem := range problems {
				fmt.Fprintln(commandStderr, problem)
			}
			c.Base().errAndQuit("The conformance corpus is stale; run `boardgame-util conformance export`")
		}
		fmt.Println("The conformance corpus is current")
		return
	}
	if c.Plain && c.Out == "" {
		c.Base().errAndQuit("--plain requires --out, so uncompressed copies are not written into the checked-in corpus")
	}
	written, err := corpus.Write(root, c.Out, c.Plain)
	if err != nil {
		c.Base().errAndQuit("Couldn't export the corpus: " + err.Error())
	}
	target := c.Out
	if target == "" {
		target = filepath.Join(root, corpus.Dir)
	}
	fmt.Printf("Wrote %d files to %s\n", len(written), target)
}

// boardgameRepoRoot walks up from the working directory to this module.
func boardgameRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "examples", "memory", "testdata", "golden")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not inside the boardgame repository; pass --root")
		}
		dir = parent
	}
}

func (c *conformanceExport) Name() string {
	return "export"
}

func (c *conformanceExport) Description() string {
	return "Regenerates the checked-in conformance corpus"
}

func (c *conformanceExport) HelpText() string {
	return c.Name() + ` replays every golden history and conformance scenario of the bundled example games on the real engine and writes one gzipped JSON document per run to conformance/ at the repository root, plus conformance/index.json.

The output is deterministic. With --check nothing is written; the command fails if a checked-in document differs from what the engine produces now.`
}

func (c *conformanceExport) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"check"}, Description: "Verify the checked-in corpus is current without writing files.", Decoder: writ.NewFlagDecoder(&c.Check), Flag: true},
		{Names: []string{"plain"}, Description: "Write uncompressed .json files for reading. Requires --out.", Decoder: writ.NewFlagDecoder(&c.Plain), Flag: true},
		{Names: []string{"out"}, Description: "Directory to write to instead of conformance/ in the repository.", Decoder: writ.NewOptionDecoder(&c.Out), Placeholder: "DIR"},
		{Names: []string{"root"}, Description: "The boardgame repository root. Defaults to the one containing the working directory.", Decoder: writ.NewOptionDecoder(&c.Root), Placeholder: "DIR"},
	}
}
