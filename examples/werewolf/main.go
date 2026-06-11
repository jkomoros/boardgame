/*
Package werewolf implements a simplified Mafia/Werewolf game. This example
demonstrates the Table+Hand companion mode with asymmetric hidden-role
information: the projector (Table view) never sees anyone's role, while
each player's phone (Hand view) shows their own role privately.
*/
package werewolf

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {
	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
}

func (g *gameDelegate) Description() string {
	return "A hidden-role game where villagers try to find the werewolves among them"
}

func (g *gameDelegate) MinNumPlayers() int {
	return 4
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 7
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 5
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	phase := game.Phase.Value()
	if phase == phaseDay || phase == phaseNight {
		// Simultaneous voting: any player may act.
		return boardgame.AnyPlayerIndex
	}
	// During gathering, no current player.
	return boardgame.ObserverPlayerIndex
}

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	_, players := concreteStates(state)

	// Initialize all player votes to -1 (no vote)
	for _, p := range players {
		p.Vote = -1
	}

	return nil
}

// Role assignment deliberately does NOT happen in FinishSetUp: at that
// point no players have been seated, and slots that never fill would
// receive roles (a werewolf on an empty seat makes the game degenerate).
// Roles are assigned by moveBeginGame at the gathering→day transition,
// among active players only. See moves.go.

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, players := concreteStates(state)

	// During gathering no roles are assigned yet, so aliveWerewolves
	// would be 0 and the "villagers win" condition would fire at game
	// creation. The game can't end before it begins.
	if game.Phase.Value() == phaseGathering {
		return false
	}

	var aliveWerewolves, aliveVillagers int

	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			aliveWerewolves++
		} else {
			aliveVillagers++
		}
	}

	// Villagers win when all werewolves are eliminated
	if aliveWerewolves == 0 {
		return true
	}

	// Werewolves win when werewolves >= villagers
	if aliveWerewolves >= aliveVillagers {
		return true
	}

	return false
}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	game, players := concreteStates(state)

	var result []string

	phase := game.Phase.Value()
	var phaseName string
	switch phase {
	case phaseGathering:
		phaseName = "Gathering"
	case phaseDay:
		phaseName = "Day"
	case phaseNight:
		phaseName = "Night"
	}
	result = append(result, fmt.Sprintf("Phase: %s  Round: %d", phaseName, game.RoundNumber+1))
	result = append(result, "")

	for i, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		line := fmt.Sprintf("Player %d", i)
		if p.Eliminated {
			line += " [ELIMINATED]"
		}
		roleName := "Villager"
		if p.Role.Value() == roleWerewolf {
			roleName = "Werewolf"
		}
		line += fmt.Sprintf(" (%s)", roleName)
		if p.Vote >= 0 {
			line += fmt.Sprintf(" -> voted for Player %d", p.Vote)
		}
		result = append(result, line)
	}

	// Check win condition
	var aliveWerewolves, aliveVillagers int
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) || p.Eliminated {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			aliveWerewolves++
		} else {
			aliveVillagers++
		}
	}
	result = append(result, "")
	result = append(result, fmt.Sprintf("Alive: %d villagers, %d werewolves", aliveVillagers, aliveWerewolves))

	return strings.Join(result, "\n")
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return &playerState{
		Vote: -1,
	}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(
				new(moves.SeatPlayer),
			),
		),
		// Gathering phase: wait for players to join.
		moves.AddOrderedForPhase(phaseGathering,
			moves.Optional(
				auto.MustConfig(
					new(moves.ActivateInactivePlayer),
					moves.WithMoveNameSuffix("Gathering"),
				),
			),
			auto.MustConfig(
				new(moves.WaitForEnoughPlayers),
				moves.WithMoveNameSuffix("Gathering"),
			),
			moves.Optional(
				auto.MustConfig(
					new(moves.InactivateEmptySeat),
					moves.WithMoveNameSuffix("Gathering"),
				),
			),
			auto.MustConfig(
				new(moveBeginGame),
				moves.WithMoveName("Begin Game"),
				moves.WithPhaseToStart(phaseDay, phaseEnum),
				moves.WithHelpText("Starts the day phase and assigns roles among seated players."),
			),
		),
		// Day and night phases share the same moves.
		// No ForceFinishTurn: werewolf uses simultaneous voting
		// (AnyPlayerIndex), so there's no "current player" whose turn
		// the host can skip. ForceFinishTurn requires CurrentPlayerSetter
		// which doesn't apply to simultaneous games.
		moves.AddForPhase(phaseDay,
			auto.MustConfig(
				new(moveCastVote),
				moves.WithHelpText("Vote for a player to eliminate."),
			),
			auto.MustConfig(
				new(moveResolveVotes),
				moves.WithHelpText("Tallies votes, eliminates the chosen player, and transitions to the next phase."),
			),
		),
		moves.AddForPhase(phaseNight,
			auto.MustConfig(
				new(moveCastVote),
				moves.WithMoveName("Cast Night Vote"),
				moves.WithHelpText("Werewolves choose a target to eliminate."),
			),
			auto.MustConfig(
				new(moveResolveVotes),
				moves.WithMoveName("Resolve Night Votes"),
				moves.WithHelpText("Tallies werewolf votes, eliminates the target, and transitions to day."),
			),
		),
	)
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	// Werewolf has no decks/components.
	return nil
}

// NewDelegate is the primary entry point of the package.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
