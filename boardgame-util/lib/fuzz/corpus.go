package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance/corpus"
)

// CorpusDir is the checked-in fuzz corpus, relative to the repository root.
const CorpusDir = "conformance/fuzz"

// CorpusIndexFile is the corpus summary, relative to CorpusDir.
const CorpusIndexFile = "index.json"

// CorpusSeeds is how many seeds each game has in the checked-in corpus.
const CorpusSeeds = 8

// Seating filters for CorpusGame.Seating.
const (
	SeatingAny      = ""
	SeatingSeated   = "seated"
	SeatingUnseated = "unseated"
)

// CorpusGame is one game of the checked-in corpus and its bounds. The bounds
// keep each game's eight documents near the size of its part of the
// existing conformance corpus; a document holds every viewer at every
// version, so a game with more viewers gets fewer versions.
type CorpusGame struct {
	Game         string `json:"game"`
	MaxProposals int    `json:"maxProposals"`
	MaxVersions  int    `json:"maxVersions"`
	// Seating, when set, keeps only seeds whose derived setup is seated (or
	// unseated): the corpus takes the first CorpusSeeds seeds from 0 up
	// that match. It selects on the setup the seed derives, never on how
	// the run turns out. A game gets one when the other setup is a dead end
	// in Go (see conformance/fuzz/README.md).
	Seating string `json:"seating,omitempty"`
}

// CorpusGames are the six games with a TypeScript port, in order.
var CorpusGames = []CorpusGame{
	{Game: "blackjack", MaxProposals: 40, MaxVersions: 70},
	// Seated, debuganimations' players stay inactive forever (it configures
	// Seat Player but no Activate Inactive Player), so nothing is legal.
	{Game: "debuganimations", MaxProposals: 30, MaxVersions: 20, Seating: SeatingUnseated},
	{Game: "memory", MaxProposals: 30, MaxVersions: 32},
	{Game: "pig", MaxProposals: 40, MaxVersions: 50},
	{Game: "tictactoe", MaxProposals: 60, MaxVersions: 60},
	// Unseated, every werewolf vote fails "your seat is not yet filled".
	{Game: "werewolf", MaxProposals: 60, MaxVersions: 80, Seating: SeatingSeated},
}

// CorpusConfigs lists every checked-in run, in file order.
func CorpusConfigs() ([]Config, error) {
	var result []Config
	for _, game := range CorpusGames {
		seeds, err := corpusSeeds(game)
		if err != nil {
			return nil, err
		}
		for _, seed := range seeds {
			result = append(result, Config{Game: game.Game, Seed: seed, Policy: PolicyRandomLegal, MaxProposals: game.MaxProposals, MaxVersions: game.MaxVersions})
		}
	}
	return result, nil
}

// corpusSeeds is the first CorpusSeeds seeds from 0 matching game.Seating.
func corpusSeeds(game CorpusGame) ([]int64, error) {
	var seeds []int64
	if game.Seating == SeatingAny {
		for seed := int64(0); seed < CorpusSeeds; seed++ {
			seeds = append(seeds, seed)
		}
		return seeds, nil
	}
	newDelegate := Delegate(game.Game)
	if newDelegate == nil {
		return nil, fmt.Errorf("unknown corpus game %q", game.Game)
	}
	manager, err := newManager(newDelegate, &driverStorage{StorageManager: newRecordStorage()})
	if err != nil {
		return nil, err
	}
	defer manager.Internals().Close()
	for seed := int64(0); len(seeds) < CorpusSeeds; seed++ {
		if seed > 1000 {
			return nil, fmt.Errorf("%s: fewer than %d seeds below 1000 are %s", game.Game, CorpusSeeds, game.Seating)
		}
		derived, err := Derive(manager, seed, seatMoveName(manager) != "")
		if err != nil {
			return nil, err
		}
		if derived.Seated == (game.Seating == SeatingSeated) {
			seeds = append(seeds, seed)
		}
	}
	return seeds, nil
}

// CorpusFile is a run's path relative to CorpusDir.
func CorpusFile(cfg Config) string {
	return Name(cfg.Game, cfg.Seed) + ".json.gz"
}

// RunOutput is one exported run, reduced to what the index needs plus,
// until a sink has consumed it, the document's bytes.
type RunOutput struct {
	Config Config
	Result *Result
	// Moves is the document's move table.
	Moves      []conformance.MoveType
	Finished   bool
	Versions   int
	Viewers    int
	Rejections int
	// Bytes is the document as Marshal renders it, uncompressed.
	Bytes []byte
}

// RunAll runs and exports every config, one after another, and returns the
// outputs in the order given. Runs are sequential on purpose: the engine and
// the example games are not safe for concurrent managers (delegates memoize
// their name in a package variable, for one). If sink is non-nil it is
// called once per output and the output's Bytes are dropped afterwards, so a
// long transient run does not hold every document in memory.
func RunAll(configs []Config, sink func(RunOutput) error) ([]RunOutput, error) {
	outputs := make([]RunOutput, len(configs))
	for i, cfg := range configs {
		out := RunOutput{Config: cfg}
		doc, result, err := Document(cfg)
		if err == nil {
			out.Result, out.Moves, out.Finished = result, doc.Moves, doc.Game.Finished
			out.Versions, out.Viewers, out.Rejections = len(doc.Versions), len(doc.Viewers), len(doc.Rejections)
			out.Bytes, err = doc.Marshal()
		}
		if err == nil && sink != nil {
			err = sink(out)
			out.Bytes = nil
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", Name(cfg.Game, cfg.Seed), err)
		}
		outputs[i] = out
	}
	return outputs, nil
}

// Index is conformance/fuzz/index.json.
type Index struct {
	Note  string          `json:"note"`
	Games []GameSummary   `json:"games"`
	Files []IndexDocument `json:"files"`
}

// IndexDocument is one document's row.
type IndexDocument struct {
	File       string   `json:"file"`
	Game       string   `json:"game"`
	Seed       int64    `json:"seed"`
	Players    int      `json:"players"`
	Variant    []string `json:"variant"`
	Seated     bool     `json:"seated"`
	Versions   int      `json:"versions"`
	Viewers    int      `json:"viewers"`
	Finished   bool     `json:"finished"`
	End        string   `json:"end"`
	Proposals  int      `json:"proposals"`
	Rejections int      `json:"rejections"`
}

// GameSummary aggregates one game's runs.
type GameSummary struct {
	Game         string `json:"game"`
	Seeds        int    `json:"seeds"`
	MaxProposals int    `json:"maxProposals"`
	MaxVersions  int    `json:"maxVersions"`
	Finished     int    `json:"finished"`
	// Moves has one row per move type, in move-table order.
	Moves []MoveSummary `json:"moves"`
	// RejectionsByKind counts rejections by their document kind
	// (propose, timer, seat); RefusalsByReason by the driver's reason.
	RejectionsByKind map[string]int `json:"rejectionsByKind"`
	RefusalsByReason map[string]int `json:"refusalsByReason"`
	// NeverApplied lists move types no run applied.
	NeverApplied []string `json:"neverApplied"`
	// Uncovered lists fields whose only candidate was their default.
	Uncovered []string `json:"uncovered"`
}

// MoveSummary is one move type's counts across a game's runs.
type MoveSummary struct {
	Name    string `json:"name"`
	IsFixUp bool   `json:"isFixUp"`
	MoveStats
}

// Summarize builds the index from finished runs. games names the games in
// order; their bounds come from CorpusGames when listed there.
func Summarize(outputs []RunOutput) *Index {
	index := &Index{
		Note:  "Generated by `boardgame-util fuzz corpus`; see README.md. Counts are summed over each game's seeds.",
		Games: []GameSummary{},
		Files: []IndexDocument{},
	}
	byGame := map[string]*GameSummary{}
	moveRows := map[string]map[string]*MoveSummary{}
	uncovered := map[string]map[string]bool{}
	for _, output := range outputs {
		cfg, result := output.Config, output.Result
		summary := byGame[cfg.Game]
		if summary == nil {
			index.Games = append(index.Games, GameSummary{
				Game: cfg.Game, MaxProposals: result.Config.MaxProposals, MaxVersions: result.Config.MaxVersions,
				Moves: []MoveSummary{}, RejectionsByKind: map[string]int{}, RefusalsByReason: map[string]int{},
				NeverApplied: []string{}, Uncovered: []string{},
			})
			summary = &index.Games[len(index.Games)-1]
			byGame[cfg.Game] = summary
			moveRows[cfg.Game] = map[string]*MoveSummary{}
			uncovered[cfg.Game] = map[string]bool{}
			for _, move := range output.Moves {
				summary.Moves = append(summary.Moves, MoveSummary{Name: move.Name, IsFixUp: move.IsFixUp})
			}
		}
		summary.Seeds++
		if output.Finished {
			summary.Finished++
		}
		for name, stats := range result.Moves {
			for i := range summary.Moves {
				if summary.Moves[i].Name == name {
					summary.Moves[i].Applied += stats.Applied
					summary.Moves[i].ProposedLegally += stats.ProposedLegally
					summary.Moves[i].Refused += stats.Refused
				}
			}
		}
		for _, refusal := range result.Refusals {
			summary.RejectionsByKind[refusal.Kind]++
			summary.RefusalsByReason[refusal.Reason]++
		}
		for _, field := range result.Uncovered {
			uncovered[cfg.Game][field] = true
		}
		index.Files = append(index.Files, IndexDocument{
			File: CorpusFile(cfg), Game: cfg.Game, Seed: cfg.Seed, Players: result.Derived.NumPlayers,
			Variant: sortedVariant(result.Derived.Variant), Seated: result.Derived.Seated,
			Versions: output.Versions, Viewers: output.Viewers, Finished: output.Finished,
			End: result.End, Proposals: result.Proposals, Rejections: output.Rejections,
		})
	}
	for i := range index.Games {
		summary := &index.Games[i]
		for _, move := range summary.Moves {
			if move.Applied == 0 {
				summary.NeverApplied = append(summary.NeverApplied, move.Name)
			}
		}
		for field := range uncovered[summary.Game] {
			summary.Uncovered = append(summary.Uncovered, field)
		}
		sort.Strings(summary.Uncovered)
	}
	return index
}

// Text renders the index's per-game summary for a terminal.
func (index *Index) Text() string {
	var out strings.Builder
	for _, game := range index.Games {
		fmt.Fprintf(&out, "%s: %d seeds, %d finished (max proposals %d, max versions %d)\n", game.Game, game.Seeds, game.Finished, game.MaxProposals, game.MaxVersions)
		fmt.Fprintf(&out, "  %-48s %8s %9s %8s\n", "move", "applied", "proposed", "refused")
		for _, move := range game.Moves {
			name := move.Name
			if move.IsFixUp {
				name += " (fix-up)"
			}
			fmt.Fprintf(&out, "  %-48s %8d %9d %8d\n", name, move.Applied, move.ProposedLegally, move.Refused)
		}
		fmt.Fprintf(&out, "  rejections by kind: %s\n", countText(game.RejectionsByKind))
		fmt.Fprintf(&out, "  refusals by reason: %s\n", countText(game.RefusalsByReason))
		if len(game.NeverApplied) > 0 {
			fmt.Fprintf(&out, "  never applied: %s\n", strings.Join(game.NeverApplied, ", "))
		}
		if len(game.Uncovered) > 0 {
			fmt.Fprintf(&out, "  default-only fields: %s\n", strings.Join(game.Uncovered, ", "))
		}
	}
	return out.String()
}

func countText(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", key, counts[key]))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// GenerateCorpus runs configs and returns the files they produce (paths
// relative to CorpusDir, uncompressed bytes) plus the index. The index is
// included only when withIndex is set.
func GenerateCorpus(configs []Config, withIndex bool) (map[string][]byte, *Index, error) {
	outputs, err := RunAll(configs, nil)
	if err != nil {
		return nil, nil, err
	}
	files := map[string][]byte{}
	for _, output := range outputs {
		files[CorpusFile(output.Config)] = output.Bytes
	}
	index := Summarize(outputs)
	if withIndex {
		blob, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return nil, nil, err
		}
		files[CorpusIndexFile] = append(blob, '\n')
	}
	return files, index, nil
}

// WriteCorpus regenerates the whole checked-in corpus under root, removing
// documents that are no longer part of it.
func WriteCorpus(root string) (*Index, error) {
	configs, err := CorpusConfigs()
	if err != nil {
		return nil, err
	}
	files, index, err := GenerateCorpus(configs, true)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, filepath.FromSlash(CorpusDir))
	for name, blob := range files {
		if strings.HasSuffix(name, ".gz") {
			if blob, err = corpus.Compress(blob); err != nil {
				return nil, err
			}
		}
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, blob, 0644); err != nil {
			return nil, err
		}
	}
	onDisk, err := filepath.Glob(filepath.Join(dir, "*", "*.json.gz"))
	if err != nil {
		return nil, err
	}
	for _, path := range onDisk {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil, err
		}
		if _, ok := files[filepath.ToSlash(rel)]; !ok {
			if err := os.Remove(path); err != nil {
				return nil, err
			}
		}
	}
	return index, nil
}

// CheckCorpus regenerates configs in memory and compares them with the
// files under root. With all set, configs must be the whole corpus: the
// index is compared too, and a checked-in document outside configs is a
// problem. It returns one line per problem; empty means current.
func CheckCorpus(root string, configs []Config, all bool) ([]string, *Index, error) {
	files, index, err := GenerateCorpus(configs, all)
	if err != nil {
		return nil, nil, err
	}
	dir := filepath.Join(root, filepath.FromSlash(CorpusDir))
	var problems []string
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if os.IsNotExist(err) {
			problems = append(problems, name+": missing")
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if strings.HasSuffix(name, ".gz") {
			if got, err = corpus.Decompress(got); err != nil {
				problems = append(problems, name+": not valid gzip: "+err.Error())
				continue
			}
		}
		if !bytes.Equal(got, want) {
			problems = append(problems, name+": differs from the engine's current output"+firstDifference(got, want))
		}
	}
	if all {
		onDisk, err := filepath.Glob(filepath.Join(dir, "*", "*.json.gz"))
		if err != nil {
			return nil, nil, err
		}
		for _, path := range onDisk {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil, nil, err
			}
			if _, ok := files[filepath.ToSlash(rel)]; !ok {
				problems = append(problems, filepath.ToSlash(rel)+": checked in but no longer part of the corpus")
			}
		}
	}
	sort.Strings(problems)
	return problems, index, nil
}

func firstDifference(got, want []byte) string {
	gotLines, wantLines := bytes.Split(got, []byte("\n")), bytes.Split(want, []byte("\n"))
	for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
		if !bytes.Equal(gotLines[i], wantLines[i]) {
			return fmt.Sprintf(" (first at line %d:\n  checked in: %.200s\n  engine:     %.200s)", i+1, gotLines[i], wantLines[i])
		}
	}
	return fmt.Sprintf(" (%d lines checked in, %d from the engine)", len(gotLines), len(wantLines))
}

// WriteRuns runs configs and writes each document, gzipped, under outDir
// as <game>/seed-NNNN.json.gz as soon as it is done, then a summary
// index.json. It is the transient run. With withResults, it also writes
// <game>/seed-NNNN.result.json: the driver's Result without the history
// (refusal reasons, per-move counts), for analysis.
func WriteRuns(configs []Config, outDir string, withResults bool) (*Index, error) {
	write := func(name string, blob []byte) error {
		path := filepath.Join(outDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, blob, 0644)
	}
	outputs, err := RunAll(configs, func(out RunOutput) error {
		blob, err := corpus.Compress(out.Bytes)
		if err != nil {
			return err
		}
		if err := write(CorpusFile(out.Config), blob); err != nil {
			return err
		}
		if !withResults {
			return nil
		}
		result := *out.Result
		result.History = nil
		blob, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		return write(Name(out.Config.Game, out.Config.Seed)+".result.json", append(blob, '\n'))
	})
	if err != nil {
		return nil, err
	}
	index := Summarize(outputs)
	blob, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, err
	}
	return index, write(CorpusIndexFile, append(blob, '\n'))
}
