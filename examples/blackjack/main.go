/*
Package blackjack implements a simple blackjack game. This example is
interesting because it has hidden state.
*/
package blackjack

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

const targetScore = 21
const defaultMaxRounds = 5
const variantKeyMaxRounds = "maxrounds"

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
}

func (g *gameDelegate) Description() string {
	return "Players draw cards trying to get as close to 21 as possible without going over"
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 7
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gameDelegate) Variants() boardgame.VariantConfig {
	maxRoundsValues := make(map[string]*boardgame.VariantDisplayInfo)
	for i := 1; i <= 10; i++ {
		s := strconv.Itoa(i)
		maxRoundsValues[s] = &boardgame.VariantDisplayInfo{
			Description: s + " rounds",
		}
	}
	return boardgame.VariantConfig{
		variantKeyMaxRounds: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				DisplayName: "Max Rounds",
				Description: "Number of rounds to play before the game ends",
			},
			Default: strconv.Itoa(defaultMaxRounds),
			Values:  maxRoundsValues,
		},
	}
}

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	game, _ := concreteStates(state)
	maxRounds, err := strconv.Atoi(variant[variantKeyMaxRounds])
	if err != nil {
		maxRounds = defaultMaxRounds
	}
	game.MaxRounds = maxRounds
	return nil
}

func (g *gameDelegate) ComputedPlayerProperties(player boardgame.ImmutableSubState) boardgame.PropertyCollection {
	result := g.GameDelegate.ComputedPlayerProperties(player)
	p := player.(*playerState)
	result["HandValue"] = p.HandValue()
	return result
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	game, _ := concreteStates(state)

	card := c.Values().(*playingcards.Card)

	if card.Rank.Value() == playingcards.RankJoker {
		return game.UnusedCards, nil
	}

	return game.DrawStack, nil

}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {

	game, players := concreteStates(state)

	var result []string

	result = append(result, fmt.Sprintf("Round: %d/%d", game.RoundsCompleted+1, game.MaxRounds))
	result = append(result, fmt.Sprintf("Cards left in deck: %d", game.DrawStack.NumComponents()))

	for i, player := range players {

		playerLine := fmt.Sprintf("Player %d", i)

		if boardgame.PlayerIndex(i) == game.CurrentPlayer {
			playerLine += "  *CURRENT*"
		}

		result = append(result, playerLine)

		handValue := player.HandValue()

		statusLine := fmt.Sprintf("\tValue: %d", handValue)

		if player.Eliminated {
			statusLine += " BUSTED"
		}

		if player.Stood {
			statusLine += " STOOD"
		}

		result = append(result, statusLine)

		result = append(result, "\tCards:")

		for _, c := range player.HiddenHand.Components() {
			result = append(result, "\t\t"+c.Values().(*playingcards.Card).String())
		}

		for _, c := range player.VisibleHand.Components() {
			result = append(result, "\t\t"+c.Values().(*playingcards.Card).String())
		}

		result = append(result, "")
	}

	return strings.Join(result, "\n")
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, _ := concreteStates(state)
	return game.RoundsCompleted >= game.MaxRounds
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(
				new(moves.ShuffleDiscardIntoDraw),
				moves.WithHelpText("When the draw deck is empty, shuffles the discard deck into draw deck."),
			),
			//Players may be seated at any time. Because playerState also has
			//behavior.InactivePlayer, the players will be seated but inactive.
			auto.MustConfig(
				new(moves.SeatPlayer),
			),
		),
		// Gathering phase: wait for players to join.
		// The gathering panel will automatically show "Waiting for Players"
		// and a share link. The game auto-starts when MinNumPlayers are seated.
		// Uses suffixed names to avoid collision with DefaultRoundSetup in
		// phaseInitialDeal (which handles between-round player re-activation).
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
				new(moves.StartPhase),
				moves.WithMoveName("Begin Game"),
				moves.WithPhaseToStart(phaseInitialDeal, phaseEnum),
			),
		),
		moves.AddForPhase(phaseNormalPlay,
			auto.MustConfig(
				new(moveStartRoundCleanup),
				moves.WithHelpText("When all players have finished, transitions to round cleanup."),
				// moves.WithPhaseToStart is gone: moveStartRoundCleanup no
				// longer embeds moves.StartPhase (see moves.go's doc
				// comment) and hand-rolls its own phase transition in
				// Apply(). moves.WithIsFixUp(true) replaces what embedding
				// moves.FixUp/moves.StartPhase provided implicitly (moves.
				// Default defaults IsFixUp to false).
				moves.WithIsFixUp(true),
				// Declarative migration (design spec §8's second flagship
				// acid test): Legal() is deleted (see moves.go); this plan
				// replaces it exactly. Default's own inPhase(phaseNormalPlay)
				// atom (contributed automatically by moves.AddForPhase's
				// WithLegalPhases call) runs first, matching the legacy
				// super-call to StartPhase.Legal -> Default.Legal.
				moves.WithPreconditions(
					legal.AllActivePlayers(
						legal.Any(legal.PlayerBool("Eliminated"), legal.PlayerBool("Stood")),
					).WithMessage("cleanup.players_unfinished"),
				),
			),
			auto.MustConfig(
				new(moveCurrentPlayerHit),
				moves.WithHelpText("The current player hits, drawing a card."),
			),
			auto.MustConfig(
				new(moveCurrentPlayerStand),
				moves.WithHelpText("If the current player no longer wants to draw cards, they can stand."),
			),
			auto.MustConfig(
				new(moveRevealHiddenCard),
				moves.WithHelpText("Reveals the hidden card in the user's hand"),
				moves.WithIsFixUp(true),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
				moves.WithHelpText("When the current player has either busted or decided to stand, we advance to next player."),
			),
			// ForceFinishTurn lets a host/server force-end the current
			// player's turn even when TurnDone() returns an error (e.g.
			// the player's phone dropped mid-turn in companion mode). Only
			// AdminPlayerIndex can propose it; FinishTurn is the normal
			// player-proposed variant. The name override is necessary
			// because ForceFinishTurn embeds FinishTurn, and the
			// auto-configurator's default name derivation would yield
			// "Finish Turn" — clashing with the parent move's registration.
			// WithIsFixUp(false) is also load-bearing: ForceFinishTurn
			// embeds FinishTurn which is a FixUp. If we let the auto-
			// proposer treat it as a fixup, the game's fixup loop would
			// propose ForceFinishTurn every turn forever (Legal always
			// returns nil for AdminPlayerIndex, the same identity used by
			// fixup proposers). The move is host-initiated only.
			auto.MustConfig(
				new(moves.ForceFinishTurn),
				moves.WithMoveName("Force Finish Turn"),
				moves.WithIsFixUp(false),
				moves.WithHelpText("Admin-only: end the current player's turn even if TurnDone() would refuse. Used by host SkipTurn in Table+Hand mode."),
			),
		),
		moves.AddOrderedForPhase(phaseInitialDeal,
			//Because we have behavior.InactivePlayer, we need to re-activate players... if there are any to run
			moves.DefaultRoundSetup(auto),
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Hidden Card"),
				moves.WithHelpText("Deals a hidden card to each player"),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("HiddenHand"),
			),
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Visible Card"),
				moves.WithHelpText("Deals a visible card to each player"),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("VisibleHand"),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(phaseNormalPlay, phaseEnum),
			),
		),
		moves.AddOrderedForPhase(phaseRoundCleanup,
			auto.MustConfig(
				new(moveAccumulateScores),
				moves.WithHelpText("Adds each player's hand value to their total score."),
			),
			auto.MustConfig(
				new(moveCollectCards),
				moves.WithHelpText("Collects all cards from players back to the discard pile."),
			),
			auto.MustConfig(
				new(moveResetPlayerForNewRound),
				moves.WithHelpText("Resets player flags for the next round."),
			),
			auto.MustConfig(
				new(moveIncrementRoundsCompleted),
				moves.WithHelpText("Increments the round counter."),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithMoveName("Start Next Round"),
				moves.WithPhaseToStart(phaseInitialDeal, phaseEnum),
			),
		),
	)

}

// ConfigureLegalTemplates supplies the one game-specific template key
// moveStartRoundCleanup's WithPreconditions plan overrides (design spec
// §8): legal.AllActivePlayers' default message ("not every active player
// satisfies the required condition") is generic, so the "cleanup.
// players_unfinished" override carries the exact legacy string from the
// pre-migration Legal() body instead.
func (g *gameDelegate) ConfigureLegalTemplates() map[string]string {
	return map[string]string{
		"cleanup.players_unfinished": "not all active players have finished their turn",
	}
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		"cards": playingcards.NewDeck(false),
	}
}

// NewDelegate is the primary entry point of the package.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
