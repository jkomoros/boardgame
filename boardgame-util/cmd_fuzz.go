package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
)

type fuzzCmd struct {
	baseSubCommand

	FuzzRun    fuzzRun
	FuzzCorpus fuzzCorpus
}

func (c *fuzzCmd) Run(p writ.Path, positional []string) {
	exitHelp(p.Last(), errors.New("SUBCOMMAND is required"))
}

func (c *fuzzCmd) Name() string {
	return "fuzz"
}

func (c *fuzzCmd) Description() string {
	return "Plays example games from seeds on the real engine and exports conformance documents with rejections"
}

func (c *fuzzCmd) HelpText() string {
	return c.Name() + ` drives a game from a seed, choosing among the actions the engine reports legal and sometimes sending an input it must refuse, and exports the run as a conformance document (boardgame-util/lib/conformance) with the refused inputs as rejections. See boardgame-util/lib/fuzz and conformance/fuzz/README.md.`
}

func (c *fuzzCmd) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.FuzzRun,
		&c.FuzzCorpus,
	}
}

type fuzzRun struct {
	baseSubCommand

	Game         string
	Seed         string
	Seeds        string
	Policy       string
	MaxProposals string
	MaxVersions  string
	RefusalRate  string
	Out          string
	Results      bool
	Root         string
}

func (c *fuzzRun) Name() string {
	return "run"
}

func (c *fuzzRun) Description() string {
	return "Runs seeds of one game and writes their documents outside the repository's corpus"
}

func (c *fuzzRun) HelpText() string {
	return c.Name() + ` runs --seed N or --seeds A..B (inclusive) of --game and writes DIR/<game>/seed-NNNN.json.gz for each, plus DIR/index.json summarizing them. Seeds run in parallel; each document depends only on its own seed and bounds. Known games: ` + strings.Join(fuzz.Games(), ", ") + `. Policies: ` + fuzz.PolicyRandomLegal + ` (default), ` + fuzz.PolicyAgent + ` (not reproducible), and ` + fuzz.PolicyBotPrefix + `<name> for ` + strings.Join(fuzz.Bots(), ", ") + `.

A run cut short with a smaller --max-versions is a prefix of the full run, so the minimal document for a failure at version V is --max-versions V.`
}

func (c *fuzzRun) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"game"}, Description: "The game to run.", Decoder: writ.NewOptionDecoder(&c.Game), Placeholder: "GAME"},
		{Names: []string{"seed"}, Description: "One seed.", Decoder: writ.NewOptionDecoder(&c.Seed), Placeholder: "N"},
		{Names: []string{"seeds"}, Description: "An inclusive range of seeds, A..B.", Decoder: writ.NewOptionDecoder(&c.Seeds), Placeholder: "A..B"},
		{Names: []string{"policy"}, Description: "random-legal (default), bot:<name> or agent.", Decoder: writ.NewOptionDecoder(&c.Policy), Placeholder: "POLICY"},
		{Names: []string{"max-proposals"}, Description: "Stop after this many inputs, refused ones included (default 300).", Decoder: writ.NewOptionDecoder(&c.MaxProposals), Placeholder: "N"},
		{Names: []string{"max-versions"}, Description: "Stop at the first settled version at or past this (default 600).", Decoder: writ.NewOptionDecoder(&c.MaxVersions), Placeholder: "N"},
		{Names: []string{"refusal-rate"}, Description: "Probability per step of sending an input the engine must refuse (default 0.125; negative for none).", Decoder: writ.NewOptionDecoder(&c.RefusalRate), Placeholder: "P"},
		{Names: []string{"out"}, Description: "Directory to write to. Must not be inside the repository's conformance/ directory.", Decoder: writ.NewOptionDecoder(&c.Out), Placeholder: "DIR"},
		{Names: []string{"results"}, Description: "Also write each run's driver result (refusal reasons, per-move counts) as <game>/seed-NNNN.result.json.", Decoder: writ.NewFlagDecoder(&c.Results), Flag: true},
		{Names: []string{"root"}, Description: "The boardgame repository root. Defaults to the one containing the working directory.", Decoder: writ.NewOptionDecoder(&c.Root), Placeholder: "DIR"},
	}
}

func (c *fuzzRun) Run(p writ.Path, positional []string) {
	if len(positional) > 0 {
		exitHelp(p.Last(), errors.New(c.Name()+" doesn't take any positional parameters"))
	}
	if c.Game == "" {
		exitHelp(p.Last(), errors.New("--game is required"))
	}
	if (c.Seed == "") == (c.Seeds == "") {
		exitHelp(p.Last(), errors.New("exactly one of --seed and --seeds is required"))
	}
	if c.Out == "" {
		exitHelp(p.Last(), errors.New("--out is required"))
	}
	if err := refuseCorpusDir(c.Root, c.Out); err != nil {
		c.Base().errAndQuit(err.Error())
	}
	seedText := c.Seed
	if seedText == "" {
		seedText = c.Seeds
	} else if strings.Contains(seedText, "..") {
		c.Base().errAndQuit("--seed takes one seed; use --seeds for a range")
	}
	seeds, err := fuzz.ParseSeeds(seedText)
	if err != nil {
		c.Base().errAndQuit(err.Error())
	}
	cfg := fuzz.Config{Game: c.Game, Policy: c.Policy}
	if cfg.MaxProposals, err = optionalInt("max-proposals", c.MaxProposals); err != nil {
		c.Base().errAndQuit(err.Error())
	}
	if cfg.MaxVersions, err = optionalInt("max-versions", c.MaxVersions); err != nil {
		c.Base().errAndQuit(err.Error())
	}
	if c.RefusalRate != "" {
		if cfg.RefusalRate, err = strconv.ParseFloat(c.RefusalRate, 64); err != nil {
			c.Base().errAndQuit("--refusal-rate: " + err.Error())
		}
		if cfg.RefusalRate == 0 {
			cfg.RefusalRate = -1
		}
	}
	configs := make([]fuzz.Config, len(seeds))
	for i, seed := range seeds {
		configs[i] = cfg
		configs[i].Seed = seed
	}
	index, err := fuzz.WriteRuns(configs, c.Out, c.Results)
	if err != nil {
		c.Base().errAndQuit("Fuzz run failed: " + err.Error())
	}
	fmt.Print(index.Text())
	fmt.Printf("Wrote %d documents to %s\n", len(configs), c.Out)
}

func optionalInt(name, text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(text)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("--%s must be a positive integer, not %q", name, text)
	}
	return value, nil
}

// refuseCorpusDir fails when out is the repository's conformance/ directory
// or inside it: transient runs never touch the checked-in corpora.
func refuseCorpusDir(rootFlag, out string) error {
	root := rootFlag
	if root == "" {
		found, err := boardgameRepoRoot()
		if err != nil {
			// Outside the repository nothing can collide with its corpus.
			return nil
		}
		root = found
	}
	corpusDir, err := filepath.Abs(filepath.Join(root, "conformance"))
	if err != nil {
		return err
	}
	target, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(filepath.Dir(corpusDir)); err == nil {
		corpusDir = filepath.Join(resolved, "conformance")
	}
	if resolved, err := evalExisting(target); err == nil {
		target = resolved
	}
	rel, err := filepath.Rel(corpusDir, target)
	if err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
		return fmt.Errorf("--out %s is inside %s; a transient run must be written outside the checked-in corpus (use `fuzz corpus` to regenerate conformance/fuzz)", out, corpusDir)
	}
	return nil
}

// evalExisting resolves symlinks in the longest existing prefix of path.
func evalExisting(path string) (string, error) {
	suffix := ""
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			return filepath.Join(resolved, suffix), nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		suffix = filepath.Join(filepath.Base(path), suffix)
		path = parent
	}
}

type fuzzCorpus struct {
	baseSubCommand

	Check bool
	Root  string
}

func (c *fuzzCorpus) Name() string {
	return "corpus"
}

func (c *fuzzCorpus) Description() string {
	return "Regenerates or checks the checked-in fuzz corpus in conformance/fuzz"
}

func (c *fuzzCorpus) HelpText() string {
	return c.Name() + ` runs the checked-in fuzz configurations (eight seeds of each ported game, with per-game bounds) and writes conformance/fuzz/<game>/seed-NNNN.json.gz plus conformance/fuzz/index.json, then prints a per-game summary: for each move type the times applied, proposed legally and refused, and the rejections by kind.

With --check nothing is written; the command fails if any checked-in document or the index differs from what the engine produces now (decompressed bytes are compared).`
}

func (c *fuzzCorpus) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"check"}, Description: "Verify the checked-in fuzz corpus is current without writing files.", Decoder: writ.NewFlagDecoder(&c.Check), Flag: true},
		{Names: []string{"root"}, Description: "The boardgame repository root. Defaults to the one containing the working directory.", Decoder: writ.NewOptionDecoder(&c.Root), Placeholder: "DIR"},
	}
}

func (c *fuzzCorpus) Run(p writ.Path, positional []string) {
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
	configs, err := fuzz.CorpusConfigs()
	if err != nil {
		c.Base().errAndQuit(err.Error())
	}
	if c.Check {
		problems, index, err := fuzz.CheckCorpus(root, configs, true)
		if err != nil {
			c.Base().errAndQuit("Couldn't generate the fuzz corpus: " + err.Error())
		}
		if len(problems) > 0 {
			for _, problem := range problems {
				fmt.Fprintln(commandStderr, problem)
			}
			c.Base().errAndQuit("The fuzz corpus is stale; run `boardgame-util fuzz corpus`")
		}
		fmt.Print(index.Text())
		fmt.Printf("The fuzz corpus is current (%d documents)\n", len(configs))
		return
	}
	index, err := fuzz.WriteCorpus(root)
	if err != nil {
		c.Base().errAndQuit("Couldn't write the fuzz corpus: " + err.Error())
	}
	fmt.Print(index.Text())
	fmt.Printf("Wrote %d documents and %s to %s\n", len(configs), fuzz.CorpusIndexFile, filepath.Join(root, filepath.FromSlash(fuzz.CorpusDir)))
}
