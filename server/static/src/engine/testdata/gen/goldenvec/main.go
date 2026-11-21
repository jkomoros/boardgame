// Command goldenvec emits ../../golden/<game>.golden.json for the example
// games whose golden histories the engine replays. For every version of every
// golden history it records what real Go says about that version:
//
//   - a digest of the full state, as storage/filesystem/record expands it
//     (so the engine's own expansion of the golden file can be checked);
//   - for every viewer (admin, observer, each player), a digest of the view
//     Game.JSONForPlayer gives that viewer, which is state.SanitizedForPlayer
//     plus the viewer JSON encoding;
//   - the full views at a few sampled versions, for diagnosis.
//
// Digests are SHA-1 over a canonical encoding defined in canon below and
// mirrored in ../../golden.test.ts. "Computed" is removed from views first:
// computed properties are game rules, outside the engine core.
//
// This is an interim stand-in for the per-viewer exporter planned under
// boardgame-util/lib/conformance. Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/goldenvec
package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

type history struct {
	// File is the golden file, relative to the repository root.
	File       string
	GameID     string
	SecretSalt string
	NumPlayers int
	Version    int
	// StateDigests[v] is the digest of the unsanitized state at version v.
	StateDigests []string
	// ViewDigests[v][viewer] is the digest of that viewer's view at v.
	// Viewers are "-2" (admin), "-1" (observer) and each player index.
	ViewDigests []map[string]string
	// Views holds full views for sampled versions: version -> viewer -> view.
	Views map[string]map[string]json.RawMessage
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// canon writes the canonical encoding of a decoded JSON value: n, t, f,
// i<decimal>; for integers, s<byte length>:<bytes> for strings, [ ... ] for
// arrays, and { key value ... } with keys in byte order for objects.
func canon(b *bytes.Buffer, v interface{}) {
	switch t := v.(type) {
	case nil:
		b.WriteString("n")
	case bool:
		if t {
			b.WriteString("t")
		} else {
			b.WriteString("f")
		}
	case json.Number:
		s := t.String()
		if strings.ContainsAny(s, ".eE") {
			panic("non-integer number in state: " + s)
		}
		b.WriteString("i" + s + ";")
	case string:
		b.WriteString("s" + strconv.Itoa(len(t)) + ":" + t)
	case []interface{}:
		b.WriteString("[")
		for _, item := range t {
			canon(b, item)
		}
		b.WriteString("]")
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("{")
		for _, k := range keys {
			canon(b, k)
			canon(b, t[k])
		}
		b.WriteString("}")
	default:
		panic(fmt.Sprintf("unexpected type %T", v))
	}
}

func decode(blob []byte) map[string]interface{} {
	dec := json.NewDecoder(bytes.NewReader(blob))
	dec.UseNumber()
	var out map[string]interface{}
	check(dec.Decode(&out))
	return out
}

func digest(v interface{}) string {
	var b bytes.Buffer
	canon(&b, v)
	sum := sha1.Sum(b.Bytes())
	return hex.EncodeToString(sum[:])
}

func process(delegate boardgame.GameDelegate, file string, first bool) history {
	rec, err := record.New(file)
	check(err)
	gameRec := rec.Game()

	storage := memstore.NewStorageManager()
	manager, err := boardgame.NewGameManager(delegate, storage)
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)

	h := history{
		File: file, GameID: gameRec.ID, SecretSalt: gameRec.SecretSalt, NumPlayers: gameRec.NumPlayers,
		Version: gameRec.Version, Views: map[string]map[string]json.RawMessage{},
	}

	for v := 0; v <= gameRec.Version; v++ {
		blob, err := rec.State(v)
		check(err)
		h.StateDigests = append(h.StateDigests, digest(decode(blob)))
		at := *gameRec
		at.Version = v
		check(storage.SaveGameAndCurrentState(&at, blob, nil))
	}

	game := manager.Game(gameRec.ID)
	if game == nil {
		panic("game did not load: " + file)
	}
	// Full views are bulky: several versions for a game's first history,
	// only the last version for the rest. Every version has digests.
	sampled := map[int]bool{gameRec.Version: true}
	if first {
		sampled = map[int]bool{0: true, 1: true, gameRec.Version / 2: true, gameRec.Version: true}
	}
	for v := 0; v <= gameRec.Version; v++ {
		state := game.State(v)
		if state == nil {
			panic(fmt.Sprintf("%s: state %d did not load", file, v))
		}
		digests := map[string]string{}
		for viewer := -2; viewer < gameRec.NumPlayers; viewer++ {
			obj, err := game.JSONForPlayer(boardgame.PlayerIndex(viewer), state)
			check(err)
			blob, err := json.Marshal(obj)
			check(err)
			var parsed struct{ CurrentState json.RawMessage }
			check(json.Unmarshal(blob, &parsed))
			view := decode(parsed.CurrentState)
			delete(view, "Computed")
			key := strconv.Itoa(viewer)
			digests[key] = digest(view)
			if sampled[v] {
				full, err := json.Marshal(view)
				check(err)
				vk := strconv.Itoa(v)
				if h.Views[vk] == nil {
					h.Views[vk] = map[string]json.RawMessage{}
				}
				h.Views[vk][key] = full
			}
		}
		h.ViewDigests = append(h.ViewDigests, digests)
	}
	return h
}

func main() {
	games := []struct {
		name     string
		delegate func() boardgame.GameDelegate
	}{
		{"memory", func() boardgame.GameDelegate { return memory.NewDelegate() }},
		{"blackjack", func() boardgame.GameDelegate { return blackjack.NewDelegate() }},
	}
	for _, g := range games {
		files, err := filepath.Glob("examples/" + g.name + "/testdata/golden/*.json")
		check(err)
		sort.Strings(files)
		var out struct{ Histories []history }
		for i, file := range files {
			out.Histories = append(out.Histories, process(g.delegate(), file, i == 0))
		}
		blob, err := json.Marshal(out)
		check(err)
		check(os.WriteFile("server/static/src/engine/testdata/golden/"+g.name+".golden.json", append(blob, '\n'), 0644))
		versions := 0
		for _, h := range out.Histories {
			versions += h.Version + 1
		}
		fmt.Fprintln(os.Stderr, g.name, "histories:", len(out.Histories), "versions:", versions)
	}
}
