package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/boardgame-util/lib/port"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

type portCmd struct {
	baseSubCommand

	PortCheck  portCheck
	PortAccept portAccept
	PortBrief  portBrief

	DryRun bool
	All    bool
	Adopt  string
	NoGame bool
	Check  bool
	portCommon
}

// portCommon are the options every port command takes.
type portCommon struct {
	Root    string
	Node    string
	Commit  string
	Profile string
}

func (c *portCommon) options() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"root"}, Description: "The boardgame repository root. Defaults to the one containing the working directory.", Decoder: writ.NewOptionDecoder(&c.Root), Placeholder: "DIR"},
		{Names: []string{"node"}, Description: "The node executable that runs the TypeScript half (Node 22.18 or later). Defaults to $BOARDGAME_NODE, else node on PATH.", Decoder: writ.NewOptionDecoder(&c.Node), Placeholder: "PATH"},
		{Names: []string{"commit"}, Description: "The commit recorded for blocks written now. Defaults to the repository's HEAD.", Decoder: writ.NewOptionDecoder(&c.Commit), Placeholder: "SHA"},
		{Names: []string{"profile"}, Description: "A text Go block profile of the fuzz corpus to read coverage from, instead of building one (fuzz coverage --blocks --work DIR keeps one).", Decoder: writ.NewOptionDecoder(&c.Profile), Placeholder: "FILE"},
	}
}

func (c *portCommon) resolve(base *boardgameUtil) (root string, node port.Node, commit string) {
	root = c.Root
	if root == "" {
		found, err := boardgameRepoRoot()
		if err != nil {
			base.errAndQuit(err.Error())
		}
		root = found
	}
	binary := c.Node
	if binary == "" {
		binary = os.Getenv("BOARDGAME_NODE")
	}
	commit = c.Commit
	if commit == "" {
		var err error
		if commit, err = headCommit(root); err != nil {
			base.errAndQuit(err.Error())
		}
	}
	return root, port.Node{Binary: binary, Root: root}, commit
}

func (c *portCmd) Name() string { return "port" }

func (c *portCmd) Description() string {
	return "Runs the Go-to-TypeScript pipeline for a game: export, translate, briefs for refusals, replay, report"
}

func (c *portCmd) HelpText() string {
	return c.Name() + ` [--dry-run] [--no-game] [--check] [--all] [--adopt FnRef] PKG...

For each game package (an import path, a directory, or a name of one of the seven example games), in order: exports its definition (gamedef), translates it (transpile), lists the refusals, writes a brief for each refused function, callee first, to ` + port.BriefsDir + `/<game>/<ts name>.md, replays the generated game (the checked-in one, else the translator's output) against the game's fuzz corpus and conformance documents on the TypeScript side, and writes the report to ` + port.ReportsDir + `/<game>.json and the table ` + port.ReportTable + `. See fabric/TRANSPILE-PROPOSAL.md, "The agent loop" and "The ratchet and its metrics".

Without --dry-run or --no-game it also writes the exported definition (when it differs from the checked-in one) and the generated game. --dry-run re-runs everything and rewrites only the reports: no generated text, brief or definition changes. --no-game writes briefs and reports but neither the definition nor the generated game. --check writes nothing and fails if a report or the table differs from what port produces now. --all covers the seven example games. --adopt FnRef replaces an agent-written block with the translator's output, after the three checks (the translator accepts the Go; with its output the corpus passes; the function's block coverage is complete).

The Go block coverage in the reports needs a profile: --profile FILE reads one, else one is built (go build -cover, then the corpus regenerated under it; about a minute). Nothing is bounded by time; child processes are waited for.

Subcommands: port check, port accept, port brief.`
}

func (c *portCmd) WritOptions() []*writ.Option {
	return append([]*writ.Option{
		{Names: []string{"dry-run"}, Description: "Re-run everything and rewrite only the reports.", Decoder: writ.NewFlagDecoder(&c.DryRun), Flag: true},
		{Names: []string{"no-game"}, Description: "Write briefs and reports, not the generated game.", Decoder: writ.NewFlagDecoder(&c.NoGame), Flag: true},
		{Names: []string{"check"}, Description: "Write nothing; fail if a checked-in report or the table is stale.", Decoder: writ.NewFlagDecoder(&c.Check), Flag: true},
		{Names: []string{"all"}, Description: "Every example game.", Decoder: writ.NewFlagDecoder(&c.All), Flag: true},
		{Names: []string{"adopt"}, Description: "Replace the agent-written block of FnRef with the translator's output, after the three checks.", Decoder: writ.NewOptionDecoder(&c.Adopt), Placeholder: "FnRef"},
	}, c.portCommon.options()...)
}

func (c *portCmd) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{&c.PortCheck, &c.PortAccept, &c.PortBrief}
}

// portGame resolves a PKG argument to one of the example games.
func portGame(arg string) (gamedefGame, bool) {
	clean := strings.TrimSuffix(strings.TrimPrefix(arg, "./"), "/")
	for _, game := range gamedefGames {
		if game.name == arg || game.importPath == arg || clean == "examples/"+game.name {
			return game, true
		}
	}
	return gamedefGame{}, false
}

// portGameOf is the example game a FnRef belongs to.
func portGameOf(ref string) (gamedefGame, bool) {
	i := strings.Index(ref, ".")
	if i <= 0 {
		return gamedefGame{}, false
	}
	return portGame(ref[:i])
}

// loadPortGame exports the game's definition and loads it for the pipeline.
func loadPortGame(base *boardgameUtil, root, commit string, game gamedefGame) (*port.Game, bool) {
	blob, note, err := exportGamedef(game)
	if note != "" {
		fmt.Fprintln(commandStderr, "Note: "+note)
	}
	if err != nil {
		base.errAndQuit("Couldn't export " + game.name + ": " + err.Error())
	}
	checkedIn, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(gamedefDir), game.name+".json"))
	current := err == nil && bytes.Equal(checkedIn, blob)
	g, err := port.Load(root, game.importPath, blob, commit)
	if err != nil {
		base.errAndQuit(err.Error())
	}
	return g, current
}

func (c *portCmd) Run(p writ.Path, positional []string) {
	base := c.Base()
	var games []gamedefGame
	if c.All {
		if len(positional) > 0 {
			exitHelp(p.Last(), errors.New("--all takes no PKG"))
		}
		for _, name := range port.ExampleGames {
			game, _ := portGame(name)
			games = append(games, game)
		}
	} else {
		if len(positional) == 0 {
			exitHelp(p.Last(), errors.New("PKG or --all is required"))
		}
		for _, arg := range positional {
			game, ok := portGame(arg)
			if !ok {
				exitHelp(p.Last(), errors.New(arg+" is not a game port knows; it knows the seven example games, by name, directory or import path"))
			}
			games = append(games, game)
		}
	}
	root, node, commit := c.resolve(base)
	writeFiles := !c.DryRun && !c.Check

	if c.Adopt != "" {
		if len(games) != 1 {
			exitHelp(p.Last(), errors.New("--adopt takes exactly one PKG"))
		}
		g, _ := loadPortGame(base, root, commit, games[0])
		result, err := g.Adopt(c.Adopt, node, c.Profile, commandStderr)
		if err != nil {
			base.errAndQuit(err.Error())
		}
		for _, problem := range result.Problems {
			fmt.Fprintln(commandStderr, problem)
		}
		if !result.Adopted {
			base.errAndQuit("Not adopted: " + c.Adopt)
		}
		fmt.Printf("Adopted the translator's output for %s in %s\n", c.Adopt, g.GeneratedPath("bodies.gen.ts"))
		return
	}

	profile := c.Profile
	if profile == "" {
		work, err := os.MkdirTemp("", "boardgame-port-profile-")
		if err != nil {
			base.errAndQuit(err.Error())
		}
		defer os.RemoveAll(work)
		fmt.Fprintln(commandStderr, "Building the Go block profile of the fuzz corpus")
		if profile, err = fuzz.BuildBlockProfile(root, work, "", nil); err != nil {
			base.errAndQuit("Couldn't measure block coverage: " + err.Error())
		}
	}

	examples := port.DefaultExamples(root)
	var stale []string
	var reports []*port.Report
	for _, game := range games {
		g, current := loadPortGame(base, root, commit, game)
		translated, refused := 0, 0
		for _, fn := range g.Output.Functions {
			if len(fn.Refusals) == 0 {
				translated++
				continue
			}
			refused++
			for _, r := range fn.Refusals {
				fmt.Printf("%s\trefused\t%s\t%s\t%s:%d\n", fn.Func, r.Reason, r.Detail, r.File, r.Line)
			}
		}
		fmt.Printf("%s: %d translated, %d refused, %d agent-written\n", g.Name, translated, refused, len(g.Output.Agent))

		for _, fn := range g.Untranslated() {
			if g.Pkg.Static.Buckets[fn.Func] == "engineGap" {
				continue
			}
			brief, err := g.WriteBrief(fn.Func, examples)
			if err != nil {
				base.errAndQuit(err.Error())
			}
			fmt.Printf("%s: brief %s (sha256 %s)\n", fn.Func, brief.Path, brief.SHA256)
			if writeFiles {
				path := filepath.Join(root, filepath.FromSlash(brief.Path))
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					base.errAndQuit(err.Error())
				}
				if err := os.WriteFile(path, brief.Text, 0644); err != nil {
					base.errAndQuit(err.Error())
				}
			}
		}

		inputs := port.ReportInputs{DefinitionCurrent: current, Profile: profile}
		var files map[string][]byte
		inputs.ReplaySource = "checked-in"
		if !g.HasGenerated() {
			files = g.Output.Files
			inputs.ReplaySource = "translator-output"
		}
		replay, err := g.Replay(node, files)
		if err != nil {
			base.errAndQuit(err.Error())
		}
		inputs.Replay = replay
		report, err := g.BuildReport(inputs)
		if err != nil {
			base.errAndQuit(err.Error())
		}
		reports = append(reports, report)
		blob, err := port.MarshalReport(report)
		if err != nil {
			base.errAndQuit(err.Error())
		}
		path := filepath.Join(root, filepath.FromSlash(port.ReportPath(g.Name)))
		if c.Check {
			if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, blob) {
				stale = append(stale, port.ReportPath(g.Name))
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				base.errAndQuit(err.Error())
			}
			if err := os.WriteFile(path, blob, 0644); err != nil {
				base.errAndQuit(err.Error())
			}
		}

		if writeFiles && !c.NoGame {
			if !current {
				path := filepath.Join(root, filepath.FromSlash(gamedefDir), g.Name+".json")
				if err := os.WriteFile(path, g.Definition, 0644); err != nil {
					base.errAndQuit(err.Error())
				}
				fmt.Printf("wrote %s\n", path)
			}
			for _, name := range transpile.FileNames {
				target := g.GeneratedPath(name)
				if got, err := os.ReadFile(target); err == nil && bytes.Equal(got, g.Output.Files[name]) {
					continue
				}
				if err := os.MkdirAll(g.GeneratedDir(), 0755); err != nil {
					base.errAndQuit(err.Error())
				}
				if err := os.WriteFile(target, g.Output.Files[name], 0644); err != nil {
					base.errAndQuit(err.Error())
				}
				fmt.Printf("wrote %s\n", target)
			}
		}
	}

	if c.All {
		table := port.TableMarkdown(reports)
		path := filepath.Join(root, filepath.FromSlash(port.ReportTable))
		if c.Check {
			if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, table) {
				stale = append(stale, port.ReportTable)
			}
		} else if err := os.WriteFile(path, table, 0644); err != nil {
			base.errAndQuit(err.Error())
		}
	}
	if c.Check {
		if len(stale) > 0 {
			for _, path := range stale {
				fmt.Fprintln(commandStderr, path+": missing or differs from what port produces now")
			}
			base.errAndQuit("The port reports are stale; run `boardgame-util port --all --dry-run`")
		}
		fmt.Println("The port reports are current")
	}
}

// ----- port check -----------------------------------------------------------

type portCheck struct {
	baseSubCommand
	File       string
	NoCoverage bool
	NoMutation bool
	portCommon
}

func (c *portCheck) Name() string { return "check" }

func (c *portCheck) Description() string {
	return "Validates an agent's translation of one function against the game's recordings"
}

func (c *portCheck) HelpText() string {
	return c.Name() + ` FnRef [--file PATH] [--no-coverage] [--no-mutation]

Checks the file (default: the brief's OUTPUT path) as the translation of FnRef: static (exactly one exported function with the required name and signature, imports only what the brief allows, the sandbox and purity rules), types (the file and the generated game with it in place type-check under src/engine/tsconfig.sandbox.json), differential (the game's checked-in fuzz corpus and conformance documents replay with the function in place; a function of the game with no translation yet is a precondition failure), then Go block coverage of the function over the fuzz corpus (unreached blocks make the result PASS (unproven)) and the body mutants of the function (survivors are listed). Prints PASS, PASS (unproven), PRECONDITION FAILED, or the first difference: version, viewer, field, JSON path, what Go recorded, what the function produced, and the arguments it was last called with. Exits nonzero unless it passes.`
}

func (c *portCheck) WritOptions() []*writ.Option {
	return append([]*writ.Option{
		{Names: []string{"file"}, Description: "The translation to check. Defaults to the brief's OUTPUT path.", Decoder: writ.NewOptionDecoder(&c.File), Placeholder: "PATH"},
		{Names: []string{"no-coverage"}, Description: "Skip the Go block coverage of the function.", Decoder: writ.NewFlagDecoder(&c.NoCoverage), Flag: true},
		{Names: []string{"no-mutation"}, Description: "Skip the body mutants.", Decoder: writ.NewFlagDecoder(&c.NoMutation), Flag: true},
	}, c.portCommon.options()...)
}

func (c *portCheck) Run(p writ.Path, positional []string) {
	result, _ := runPortCheck(c.Base(), p, positional, c.File, c.NoCoverage, c.NoMutation, &c.portCommon, "")
	fmt.Fprint(commandStdout, result.Text())
	if result.Verdict != port.VerdictPass && result.Verdict != port.VerdictUnproven {
		quit(1)
	}
}

// runPortCheck checks (or, with producer, accepts) one function.
func runPortCheck(base *boardgameUtil, p writ.Path, positional []string, file string, noCoverage, noMutation bool, common *portCommon, producer string) (*port.CheckResult, bool) {
	if len(positional) != 1 {
		exitHelp(p.Last(), errors.New("exactly one FnRef is required"))
	}
	ref := positional[0]
	game, ok := portGameOf(ref)
	if !ok {
		exitHelp(p.Last(), errors.New(ref+" is not a function of one of the seven example games"))
	}
	root, node, commit := common.resolve(base)
	g, _ := loadPortGame(base, root, commit, game)
	if file == "" {
		file = filepath.Join(root, filepath.FromSlash(port.OutputPath(g.Name, ref)))
	}
	candidate, err := os.ReadFile(file)
	if err != nil {
		base.errAndQuit(err.Error())
	}
	opts := port.CheckOptions{Node: node, Examples: port.DefaultExamples(root), Coverage: !noCoverage, Profile: common.Profile, Mutation: !noMutation, Log: commandStderr}
	if noCoverage {
		opts.Profile = ""
	}
	if producer == "" {
		result, err := g.Check(ref, candidate, opts)
		if err != nil {
			base.errAndQuit(err.Error())
		}
		return result, false
	}
	result, stored, err := g.Accept(ref, candidate, producer, opts, port.FilePairWriter{Root: root})
	if err != nil {
		base.errAndQuit(err.Error())
	}
	return result, stored
}

// ----- port accept ----------------------------------------------------------

type portAccept struct {
	baseSubCommand
	File       string
	Producer   string
	NoMutation bool
	portCommon
}

func (c *portAccept) Name() string { return "accept" }

func (c *portAccept) Description() string {
	return "Checks an agent's translation and copies it into the generated game"
}

func (c *portAccept) HelpText() string {
	return c.Name() + ` FnRef --file PATH --producer MODEL

Runs port check on the file and, if it passes (proven or unproven), copies the function into the generated game's bodies.gen.ts under the header // @generated agent:<MODEL> brief:<sha256> go:<FnRef> sha256:<hash> first:<commit>, and stores the pair in ` + port.PairsDir + ` when that store exists. The translator keeps the block for as long as the Go function's source hash is unchanged; port --adopt replaces it with the translator's output on purpose.`
}

func (c *portAccept) WritOptions() []*writ.Option {
	return append([]*writ.Option{
		{Names: []string{"file"}, Description: "The translation to accept.", Decoder: writ.NewOptionDecoder(&c.File), Placeholder: "PATH"},
		{Names: []string{"producer"}, Description: "The model id of the agent that wrote it.", Decoder: writ.NewOptionDecoder(&c.Producer), Placeholder: "MODEL"},
		{Names: []string{"no-mutation"}, Description: "Skip the body mutants.", Decoder: writ.NewFlagDecoder(&c.NoMutation), Flag: true},
	}, c.portCommon.options()...)
}

func (c *portAccept) Run(p writ.Path, positional []string) {
	if c.File == "" || c.Producer == "" {
		exitHelp(p.Last(), errors.New("--file and --producer are required"))
	}
	result, stored := runPortCheck(c.Base(), p, positional, c.File, false, c.NoMutation, &c.portCommon, c.Producer)
	fmt.Fprint(commandStdout, result.Text())
	if result.Verdict != port.VerdictPass && result.Verdict != port.VerdictUnproven {
		c.Base().errAndQuit("Not accepted")
	}
	fmt.Fprintf(commandStdout, "Accepted %s into the generated game (producer %s)\n", result.Func, c.Producer)
	if stored {
		fmt.Fprintf(commandStdout, "Stored the pair under %s\n", port.PairsDir)
	} else {
		fmt.Fprintf(commandStdout, "No pair stored: %s\n", port.ErrNoPairStore)
	}
}

// ----- port brief -----------------------------------------------------------

type portBrief struct {
	baseSubCommand
	portCommon
}

func (c *portBrief) Name() string { return "brief" }

func (c *portBrief) Description() string { return "Prints the brief for one function" }

func (c *portBrief) HelpText() string {
	return c.Name() + ` FnRef prints the brief port writes for FnRef (any function of an example game's manifest), and its SHA-256 on stderr.`
}

func (c *portBrief) WritOptions() []*writ.Option { return c.portCommon.options() }

func (c *portBrief) Run(p writ.Path, positional []string) {
	if len(positional) != 1 {
		exitHelp(p.Last(), errors.New("exactly one FnRef is required"))
	}
	game, ok := portGameOf(positional[0])
	if !ok {
		exitHelp(p.Last(), errors.New(positional[0]+" is not a function of one of the seven example games"))
	}
	root, _, commit := c.resolve(c.Base())
	g, _ := loadPortGame(c.Base(), root, commit, game)
	brief, err := g.WriteBrief(positional[0], port.DefaultExamples(root))
	if err != nil {
		c.Base().errAndQuit(err.Error())
	}
	commandStdout.Write(brief.Text)
	fmt.Fprintf(commandStderr, "sha256 %s\n", brief.SHA256)
}
