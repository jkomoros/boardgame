package bot_test

import (
	"path/filepath"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/tictactoe/bot"
)

func TestObservationBotRunsThroughScenarioAndGolden(t *testing.T) {
	spec := scenario.Spec{Name: "Tic-tac-toe bot opening", Seed: "bot-opening", Players: 2, Viewers: []boardgame.PlayerIndex{0, 1}, Steps: []scenario.Step{
		{Label: "First random placement", Player: 0, Bot: bot.RandomEmpty{}},
		{Label: "Second random placement", Player: 1, Bot: bot.RandomEmpty{}},
	}}
	result, err := scenario.Run(tictactoe.NewDelegate(), spec)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := scenario.Run(tictactoe.NewDelegate(), spec)
	if err != nil {
		t.Fatal(err)
	}
	for i, frame := range result.Replay.Frames {
		for j, view := range frame.Viewers {
			if string(view.Game) != string(repeated.Replay.Frames[i].Viewers[j].Game) {
				t.Fatal("injected RNG did not reproduce observation")
			}
		}
	}
	path := filepath.Join(t.TempDir(), "bot.json")
	if err := result.History.Save(path, false); err != nil {
		t.Fatal(err)
	}
	if err := golden.Compare(tictactoe.NewDelegate(), path, false); err != nil {
		t.Fatal(err)
	}
}
