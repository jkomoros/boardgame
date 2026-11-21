// Package corpus defines the checked-in conformance corpus: which runs of the
// example games are exported, and where the documents live.
//
// The corpus is written to conformance/ at the repository root:
//
//	conformance/schema.ts               the document format, as TypeScript types
//	conformance/index.json              inventory of every document
//	conformance/<game>/<name>.json.gz   one conformance.Document, gzipped
//
// Documents are gzipped because a full per-viewer record is large (several
// megabytes for a long game) but extremely repetitive. Compression is
// deterministic, and the staleness test compares the decompressed bytes, so
// a change in Go's compressor cannot fail it.
//
// Regenerate with:
//
//	./scripts/go-local run ./boardgame-util conformance export
package corpus

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	memoryscenarios "github.com/jkomoros/boardgame/examples/memory/scenarios"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	werewolfscenarios "github.com/jkomoros/boardgame/examples/werewolf/scenarios"
)

// Dir is the corpus directory, relative to the repository root.
const Dir = "conformance"

// IndexFile is the inventory, relative to Dir.
const IndexFile = "index.json"

// Entry is one document of the corpus.
type Entry struct {
	// File is the document's path relative to Dir, ending in .json.gz.
	File string
	// Build exports the document. root is the repository root.
	Build func(root string) (*conformance.Document, error)
}

// IndexEntry is one row of index.json.
type IndexEntry struct {
	File     string `json:"file"`
	Game     string `json:"game"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Versions int    `json:"versions"`
	Viewers  int    `json:"viewers"`
	Finished bool   `json:"finished"`
	Covers   string `json:"covers"`
}

var games = map[string]conformance.NewDelegate{
	"blackjack":       func() boardgame.GameDelegate { return blackjack.NewDelegate() },
	"debuganimations": func() boardgame.GameDelegate { return debuganimations.NewDelegate() },
	"memory":          func() boardgame.GameDelegate { return memory.NewDelegate() },
	"pig":             func() boardgame.GameDelegate { return pig.NewDelegate() },
	"tictactoe":       func() boardgame.GameDelegate { return tictactoe.NewDelegate() },
	"werewolf":        func() boardgame.GameDelegate { return werewolf.NewDelegate() },
}

// goldenGames are the games whose every golden history is exported.
var goldenGames = []string{"blackjack", "memory", "pig", "tictactoe"}

// goldenCovers notes what a golden adds; goldens not listed get a generic note.
var goldenCovers = map[string]string{
	"blackjack/82ef86840a83b9ed": "Four seats with two players seated through the seating rendezvous; empty seats inactivated; one round and the next deal.",
	"blackjack/dac0bc2b57492b3a": "Seven players, one full round with hits and stands, then the next deal.",
	"memory/6cb9690b3b55fe3d":    "A complete game to game-over with a winner: captures, timer-fired hides and a player-proposed hide.",
	"pig/67672ddc7e8a71ad":       "A complete game to game-over with a winner; every die roll comes from State.Rand.",
	"tictactoe/1b7c0a299f5b79c5": "A complete game ending in a draw: finished with no winners.",
	"tictactoe/0e7f8ddf5c1c7c65": "A complete game won by player 0.",
	"tictactoe/568a53942f7d083d": "A complete game won by player 1.",
}

type scenarioEntry struct {
	game, file, covers string
	spec               func() (scenario.Spec, error)
}

var scenarios = []scenarioEntry{
	{"memory", "scenario-mismatch", "The existing review scenario: two different cards revealed, the hide-cards timer started by a fix-up and then fired.", memoryscenarios.Mismatch},
	{"werewolf", "scenario-timed-vote", "The existing review scenario: four players seated, role assignment hidden per player, one vote whose target only the voter and same-role players may see, and a deadline timer resolving the vote.", werewolfscenarios.TimedVote},
	{"memory", "scenario-full-game", "A small game played to game-over with winners: a non-match hidden by the timer, a non-match hidden by the player (cancelling the timer), then every pair captured.", memoryFullGame},
	{"blackjack", "scenario-bust-and-finish", "Three players, two rounds, to game-over with winners: a bust, stands, the hidden card revealed at each turn (len-policy hands seen by self, other, observer and admin), round cleanup and the redeal.", blackjackBustAndFinish},
	{"blackjack", "scenario-reshuffle", "Seven players standing through four rounds to game-over: the 52-card draw stack is exhausted during the fourth deal, so Shuffle Discard Into Draw fires mid-progression (a shuffle after set-up, from that version's seeded randomness, raising SecretMoveCount and re-scrambling ids of a len-policy stack).", blackjackReshuffle},
	{"debuganimations", "scenario-shuffles", "Mid-game shuffles that re-scramble component ids (secret, public, and a hidden-name shuffle whose move record other viewers do not receive at all), plus moves between order- and nonempty-policy stacks.", debugAnimationsShuffles},
	{"pig", "scenario-short", "A few turns of dice rolls from State.Rand, including a turn ended by the player.", pigShort},
	{"tictactoe", "scenario-seated-win", "Both players seated through the seating rendezvous, then a win. Place Token's name is visible only to its proposer: other viewers receive the alias animation key and no arguments.", ticTacToeSeatedWin},
}

// Entries lists the corpus in a stable order. Goldens are discovered from
// each game's testdata, so a new golden makes the corpus stale until it is
// regenerated.
func Entries(root string) ([]Entry, error) {
	var entries []Entry
	for _, game := range goldenGames {
		game := game
		pattern := filepath.Join(root, "examples", game, "testdata", "golden", "*.json")
		paths, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("no golden histories match %s", pattern)
		}
		sort.Strings(paths)
		for _, path := range paths {
			id := strings.TrimSuffix(filepath.Base(path), ".json")
			display := "examples/" + game + "/testdata/golden/" + id + ".json"
			covers := goldenCovers[game+"/"+id]
			if covers == "" {
				covers = "A recorded golden history, replayed in full."
			}
			entries = append(entries, Entry{
				File: game + "/golden-" + id + ".json.gz",
				Build: func(root string) (*conformance.Document, error) {
					return conformance.FromGolden(games[game], filepath.Join(root, filepath.FromSlash(display)), display, game+" golden "+id, covers)
				},
			})
		}
	}
	for _, entry := range scenarios {
		entry := entry
		entries = append(entries, Entry{
			File: entry.game + "/" + entry.file + ".json.gz",
			Build: func(string) (*conformance.Document, error) {
				spec, err := entry.spec()
				if err != nil {
					return nil, fmt.Errorf("build scenario: %w", err)
				}
				return conformance.FromScenario(games[entry.game], spec, entry.covers)
			},
		})
	}
	return entries, nil
}

// Generate exports the whole corpus in memory. Keys are paths relative to
// Dir; values are the uncompressed bytes (a Document, or the index).
func Generate(root string) (map[string][]byte, error) {
	entries, err := Entries(root)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte, len(entries)+1)
	index := make([]IndexEntry, 0, len(entries))
	for _, entry := range entries {
		doc, err := entry.Build(root)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.File, err)
		}
		blob, err := doc.Marshal()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.File, err)
		}
		files[entry.File] = blob
		source := doc.Source.Kind
		if doc.Source.Path != "" {
			source += " " + doc.Source.Path
		}
		index = append(index, IndexEntry{
			File: entry.File, Game: doc.Game.Name, Name: doc.Name, Source: source,
			Versions: len(doc.Versions), Viewers: len(doc.Viewers),
			Finished: doc.Game.Finished, Covers: doc.Source.Covers,
		})
	}
	blob, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, err
	}
	files[IndexFile] = append(blob, '\n')
	return files, nil
}

// Write regenerates the corpus under root/Dir, removing documents that are
// no longer part of it. With plain set it writes uncompressed .json files
// to outDir instead, for reading; those are not the checked-in form.
func Write(root, outDir string, plain bool) ([]string, error) {
	files, err := Generate(root)
	if err != nil {
		return nil, err
	}
	if outDir == "" {
		outDir = filepath.Join(root, Dir)
	}
	var written []string
	for name, blob := range files {
		if strings.HasSuffix(name, ".gz") {
			if plain {
				name = strings.TrimSuffix(name, ".gz")
			} else if blob, err = Compress(blob); err != nil {
				return nil, err
			}
		}
		path := filepath.Join(outDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, blob, 0644); err != nil {
			return nil, err
		}
		written = append(written, name)
	}
	sort.Strings(written)
	if plain {
		return written, nil
	}
	stale, err := filepath.Glob(filepath.Join(outDir, "*", "*.json.gz"))
	if err != nil {
		return nil, err
	}
	for _, path := range stale {
		rel, err := filepath.Rel(outDir, path)
		if err != nil {
			return nil, err
		}
		if _, ok := files[filepath.ToSlash(rel)]; !ok {
			if err := os.Remove(path); err != nil {
				return nil, err
			}
		}
	}
	return written, nil
}

// Compress gzips a document with no timestamp or file name in the header.
func Compress(blob []byte) ([]byte, error) {
	var out bytes.Buffer
	writer, err := gzip.NewWriterLevel(&out, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(blob); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Decompress reverses Compress.
func Decompress(blob []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// Check regenerates the corpus in memory and compares it with the files
// checked in under root/Dir. It returns one line per stale, missing or
// unexpected file; an empty result means the corpus is current.
func Check(root string) ([]string, error) {
	files, err := Generate(root)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, Dir)
	var problems []string
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if os.IsNotExist(err) {
			problems = append(problems, name+": missing")
			continue
		}
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(name, ".gz") {
			if got, err = Decompress(got); err != nil {
				problems = append(problems, name+": not valid gzip: "+err.Error())
				continue
			}
		}
		if !bytes.Equal(got, want) {
			problems = append(problems, name+": differs from the engine's current output"+firstDifference(got, want))
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
			problems = append(problems, filepath.ToSlash(rel)+": checked in but no longer part of the corpus")
		}
	}
	sort.Strings(problems)
	return problems, nil
}

// firstDifference describes where two documents first differ, by line.
func firstDifference(got, want []byte) string {
	gotLines, wantLines := bytes.Split(got, []byte("\n")), bytes.Split(want, []byte("\n"))
	for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
		if !bytes.Equal(gotLines[i], wantLines[i]) {
			return fmt.Sprintf(" (first at line %d:\n  checked in: %.200s\n  engine:     %.200s)", i+1, gotLines[i], wantLines[i])
		}
	}
	return fmt.Sprintf(" (%d lines checked in, %d from the engine)", len(gotLines), len(wantLines))
}
