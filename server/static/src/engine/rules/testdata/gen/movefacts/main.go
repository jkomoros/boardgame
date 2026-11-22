// Command movefacts emits ../../movefacts/<game>.json: for each move of each
// example game whose corpus the engine replays, the facts about it that the
// live Go manager knows and that a GameData fixture's MoveData must agree
// with: name, help text, fix-up flag, legal phases, progression (as
// moves.DescribeProgression gives it), AllowMultipleInProgression, and the
// assembled legal plan ((*GameManager).LegalPlanSpecs). rules/data.test.ts
// checks the hand-written fixtures in ../../ against these, so the parts of
// a fixture that are hardest to get right by hand are held to Go.
//
// Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/rules/testdata/gen/movefacts
package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/interfaces"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

type plan struct {
	Declared  bool
	Specs     []boardgame.LegalSpec
	HasCustom bool
}

type moveFacts struct {
	Name                       string
	HelpText                   string
	IsFixUp                    bool
	LegalPhases                []int
	Progression                *moves.ProgressionDescription
	AllowMultipleInProgression bool
	Plan                       plan
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	delegates := []boardgame.GameDelegate{
		blackjack.NewDelegate(), debuganimations.NewDelegate(), memory.NewDelegate(),
		pig.NewDelegate(), tictactoe.NewDelegate(), werewolf.NewDelegate(),
	}
	for _, delegate := range delegates {
		manager, err := boardgame.NewGameManager(delegate, memstore.NewStorageManager())
		check(err)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		manager.SetLogger(logger)
		facts := []moveFacts{}
		for _, move := range manager.ExampleMoves() {
			name := move.Info().Name()
			f := moveFacts{Name: name, HelpText: move.HelpText()}
			if fixUp, ok := move.(interface{ IsFixUp() bool }); ok {
				f.IsFixUp = fixUp.IsFixUp()
			}
			if phased, ok := move.(interface{ InstalledLegalPhases() []enum.EnumKey }); ok {
				for _, key := range phased.InstalledLegalPhases() {
					f.LegalPhases = append(f.LegalPhases, int(key))
				}
			}
			if group := moves.InstalledProgression(move.Info().CustomConfiguration()); group != nil {
				desc := moves.DescribeProgression(group)
				f.Progression = &desc
			}
			if multi, ok := move.(interfaces.AllowMultipleInProgression); ok {
				f.AllowMultipleInProgression = multi.AllowMultipleInProgression()
			}
			specs, declared, hasCustom, ok := manager.LegalPlanSpecs(name)
			if !ok {
				panic("LegalPlanSpecs does not know " + name)
			}
			if specs == nil {
				specs = []boardgame.LegalSpec{}
			}
			f.Plan = plan{Declared: declared, Specs: specs, HasCustom: hasCustom}
			facts = append(facts, f)
		}
		blob, err := json.MarshalIndent(facts, "", "\t")
		check(err)
		check(os.MkdirAll("server/static/src/engine/rules/testdata/movefacts", 0755))
		check(os.WriteFile("server/static/src/engine/rules/testdata/movefacts/"+manager.Delegate().Name()+".json", append(blob, '\n'), 0644))
	}
}
