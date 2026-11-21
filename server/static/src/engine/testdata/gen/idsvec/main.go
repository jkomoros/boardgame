// Command idsvec emits ../../ids.vectors.json: SHA-1 known answers from
// crypto/sha1, and component ids read from real boardgame states (so the
// vectors include the exact concatenation the core hashes and non-zero
// secret-move counts). Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/idsvec
package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

type shaVec struct {
	// InputHex is the input bytes, hex encoded.
	InputHex string
	Digest   string
}

type idVec struct {
	GameID          string
	SecretSalt      string
	Deck            string
	DeckIndex       int
	SecretMoveCount int
	ID              string
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var out struct {
		SHA1 []shaVec
		IDs  []idVec
	}

	// Every length across three block boundaries, plus fixed texts.
	for n := 0; n <= 200; n++ {
		buf := make([]byte, n)
		for i := range buf {
			buf[i] = byte((i*131 + n*17 + 7) % 256)
		}
		sum := sha1.Sum(buf)
		out.SHA1 = append(out.SHA1, shaVec{hex.EncodeToString(buf), hex.EncodeToString(sum[:])})
	}
	for _, text := range []string{"abc", "The quick brown fox jumps over the lazy dog", strings.Repeat("a", 1000), "héllo 🎲"} {
		sum := sha1.Sum([]byte(text))
		out.SHA1 = append(out.SHA1, shaVec{hex.EncodeToString([]byte(text)), hex.EncodeToString(sum[:])})
	}

	h, err := vecgame.NewHarness()
	check(err)

	games := []struct{ id, salt string }{
		{"FAKEIDFORTESTING", "FAKESALTFORTESTING"},
		{"6CB9690B3B55FE3D", "25F10986D87C2721"},
		{"A", "B"},
		{"0123456789ABCDEF", "FEDCBA9876543210"},
		{"ID WITH SPACES", "sält-ünïcode"},
		{"12", "3"}, // id+salt+deck+index+count has no separators
		{"1", "23"},
	}
	for round, g := range games {
		blob, err := h.FreshBlob(g.id, g.salt, 3)
		check(err)
		_, st, err := h.StateAt(g.id+"X", g.salt, 5+round, 3, blob)
		check(err)
		game := vecgame.Game(st)
		players := vecgame.Players(st)
		// Drive counts up unevenly: whole-stack shuffles, and secret moves
		// that scramble only the destination.
		for i := 0; i < round+1; i++ {
			check(game.Draw.Shuffle())
			check(game.Draw.ComponentAt(0).SecretMoveTo(players[i%3].Hand, 0))
			if i < 4 {
				check(game.Draw.ComponentAt(0).MoveTo(game.Row, i))
			}
			if i%2 == 0 {
				check(players[i%3].Hand.Shuffle())
			}
		}
		check(game.Tokens.ComponentAt(0).SecretMoveTo(game.Spaces.SpaceAt(1), 0))

		var parsed struct{ SecretMoveCount map[string][]int }
		blobAfter, err := vecgame.Blob(st)
		check(err)
		check(json.Unmarshal(blobAfter, &parsed))

		for _, deckName := range []string{"cards", "tokens"} {
			deck := h.Manager.Chest().Deck(deckName)
			for _, c := range deck.Components() {
				count := 0
				if counts := parsed.SecretMoveCount[deckName]; counts != nil {
					count = counts[c.DeckIndex()]
				}
				out.IDs = append(out.IDs, idVec{
					GameID: g.id + "X", SecretSalt: g.salt, Deck: deckName, DeckIndex: c.DeckIndex(),
					SecretMoveCount: count, ID: c.ImmutableInstance(st).ID(),
				})
			}
			// The generic component has no id.
			generic := deck.GenericComponent().ImmutableInstance(st).ID()
			if generic != "" {
				panic("generic component had an id")
			}
		}
	}

	nonZero := 0
	for _, v := range out.IDs {
		if v.SecretMoveCount > 0 {
			nonZero++
		}
	}
	if nonZero < 20 {
		panic(fmt.Sprintf("only %d ids with a non-zero count", nonZero))
	}

	blob, err := json.Marshal(out)
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/ids.vectors.json", append(blob, '\n'), 0644))
}
