package fuzz

import (
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/bots"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	tictactoebot "github.com/jkomoros/boardgame/examples/tictactoe/bot"
	"github.com/jkomoros/boardgame/examples/werewolf"
)

var games = map[string]conformance.NewDelegate{
	"blackjack":       func() boardgame.GameDelegate { return blackjack.NewDelegate() },
	"checkers":        func() boardgame.GameDelegate { return checkers.NewDelegate() },
	"debuganimations": func() boardgame.GameDelegate { return debuganimations.NewDelegate() },
	"memory":          func() boardgame.GameDelegate { return memory.NewDelegate() },
	"pig":             func() boardgame.GameDelegate { return pig.NewDelegate() },
	"tictactoe":       func() boardgame.GameDelegate { return tictactoe.NewDelegate() },
	"werewolf":        func() boardgame.GameDelegate { return werewolf.NewDelegate() },
}

// Games lists the game names the driver knows, sorted.
func Games() []string {
	names := make([]string, 0, len(games))
	for name := range games {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Delegate returns the constructor for a known game, or nil.
func Delegate(game string) conformance.NewDelegate {
	return games[game]
}

// bots are the registered seeded policies: a name, and the game it plays.
var botPolicies = map[string]struct {
	game   string
	policy func() bots.Policy
}{
	"tictactoe-random-empty": {"tictactoe", func() bots.Policy { return tictactoebot.RandomEmpty{} }},
}

// Bots lists the names accepted after PolicyBotPrefix, sorted.
func Bots() []string {
	names := make([]string, 0, len(botPolicies))
	for name := range botPolicies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
