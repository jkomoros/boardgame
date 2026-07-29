package moves

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
DrawToPlayer is "draw a card": it moves one component from a stack on gameState
into a stack belonging to the current player. It is the other of the two most
common verbs in the medium, and it was unexpressible until the component-moving
moves could reach player state at all.

It discovers as much as it can:

  - The source is the draw stack of the gameState's [behaviors.DrawDiscardPair],
    when there is one. So a game that already embeds that behavior -- the same
    one [ShuffleDiscardIntoDraw] uses -- needs no source configuration:

    auto.MustConfig(new(moves.DrawToPlayer))

  - The destination is the current player's stack. When the playerState has
    exactly one stack, that is it. Otherwise name it with [WithPlayerProperty].

[WithGameProperty] overrides the discovered source, for a game with several
piles or none of the behavior at all. Every one of those decisions is made and
checked once, at NewGameManager time.

Like [MoveComponentToSlot], this move declares no Legal() of its own: its
legality is contributed declaratively (legal.MayMoveFirstTo) on top of
[CurrentPlayer]'s checks, so an embedding move keeps its own
[WithLegalPreconditions] and LegalCustom.

boardgame:codegen
*/
type DrawToPlayer struct {
	CurrentPlayer
}

// These are not creator-facing options; they are where ValidConfiguration
// records what it discovered so ContributedPreconditions can read it back.
const configPropResolvedDrawSource = fullyQualifiedPackageName + "ResolvedDrawSource"
const configPropResolvedDrawDestination = fullyQualifiedPackageName + "ResolvedDrawDestination"

// drawPlan is the resolved source and destination property names. They are
// names, not stacks, because the contributed legal specs are path-based.
type drawPlan struct {
	gameStackName   string
	playerStackName string
}

func drawPlanFor(move boardgame.Move, state boardgame.State) (drawPlan, error) {

	infoer, ok := move.(moveInfoer)

	if !ok {
		return drawPlan{}, errors.New("move did not expose its custom configuration")
	}

	gameStackName, err := drawSourceName(infoer, state)

	if err != nil {
		return drawPlan{}, err
	}

	playerStackName, err := drawDestinationName(infoer, state)

	if err != nil {
		return drawPlan{}, err
	}

	return drawPlan{gameStackName: gameStackName, playerStackName: playerStackName}, nil
}

func drawSourceName(infoer moveInfoer, state boardgame.State) (string, error) {

	if configured, ok := infoer.CustomConfiguration()[configPropGameProperty]; ok {
		name, ok := configured.(string)
		if !ok || name == "" {
			return "", errors.New("WithGameProperty was not given a property name")
		}
		if _, err := state.GameState().ReadSetter().StackProp(name); err != nil {
			return "", fmt.Errorf("WithGameProperty named %q: %w", name, err)
		}
		return name, nil
	}

	hasPair, ok := state.GameState().(behaviors.HasDrawDiscardPair)

	if !ok {
		return "", errors.New("no draw stack: pass WithGameProperty, or embed behaviors.DrawDiscardPair on your gameState")
	}

	pair := hasPair.GetDrawDiscardPair()

	if pair == nil {
		return "", errors.New("gameState's DrawDiscardPair was nil")
	}

	name := stackPropName(state.GameState().Reader(), pair.DrawStack())

	if name == "" {
		return "", errors.New("couldn't find the DrawDiscardPair's draw stack among gameState's properties; name it with WithGameProperty")
	}

	return name, nil
}

func drawDestinationName(infoer moveInfoer, state boardgame.State) (string, error) {

	players := state.PlayerStates()

	if len(players) == 0 {
		return "", errors.New("the game has no players")
	}

	reader := players[0].ReadSetter()

	if configured, ok := infoer.CustomConfiguration()[configPropPlayerProperty]; ok {
		name, ok := configured.(string)
		if !ok || name == "" {
			return "", errors.New("WithPlayerProperty was not given a property name")
		}
		if _, err := reader.StackProp(name); err != nil {
			return "", fmt.Errorf("WithPlayerProperty named %q: %w", name, err)
		}
		return name, nil
	}

	var candidates []string
	for name, propType := range reader.Props() {
		if propType == boardgame.TypeStack {
			candidates = append(candidates, name)
		}
	}
	sort.Strings(candidates)

	switch len(candidates) {
	case 1:
		return candidates[0], nil
	case 0:
		return "", errors.New("playerState has no stack to draw into; add one, or name it with WithPlayerProperty")
	}

	return "", fmt.Errorf("playerState has more than one stack (%v), so the one to draw into is ambiguous; name it with WithPlayerProperty", candidates)
}

// stackPropName returns the raw property name of stack within reader, or "".
// findStackNameInReader (default.go) is the display-name counterpart; this one
// returns the name a property path can actually use.
func stackPropName(reader boardgame.PropertyReader, stack boardgame.ImmutableStack) string {
	if stack == nil {
		return ""
	}
	for name, propType := range reader.Props() {
		if propType != boardgame.TypeStack {
			continue
		}
		candidate, err := reader.ImmutableStackProp(name)
		if err != nil {
			continue
		}
		if candidate == stack {
			return name
		}
	}
	return ""
}

// GameStack returns the stack components are drawn from.
func (d *DrawToPlayer) GameStack(gameState boardgame.SubState) boardgame.Stack {
	plan, err := drawPlanFor(d.Info().ConcreteMove(), gameState.State())
	if err != nil {
		return nil
	}
	stack, err := gameState.ReadSetter().StackProp(plan.gameStackName)
	if err != nil {
		return nil
	}
	return stack
}

// PlayerStack returns the stack components are drawn into.
func (d *DrawToPlayer) PlayerStack(playerState boardgame.SubState) boardgame.Stack {
	plan, err := drawPlanFor(d.Info().ConcreteMove(), playerState.State())
	if err != nil {
		return nil
	}
	stack, err := playerState.ReadSetter().StackProp(plan.playerStackName)
	if err != nil {
		return nil
	}
	return stack
}

// ContributedPreconditions returns CurrentPlayer's checks plus the one atom
// that IS this move: may the top of the draw stack go into the current
// player's stack.
func (d *DrawToPlayer) ContributedPreconditions() []legal.Spec {

	specs := d.CurrentPlayer.ContributedPreconditions()

	//ContributedPreconditions is handed no state, but the plan is only
	//property NAMES, which every state agrees on -- so ValidConfiguration
	//resolves it once against the example state and records it here.
	//NewGameManager runs every move's ValidConfiguration before it assembles
	//any legal plan, so the record is always present by the time this runs.
	config := d.CustomConfiguration()
	gameStackName, _ := config[configPropResolvedDrawSource].(string)
	playerStackName, _ := config[configPropResolvedDrawDestination].(string)

	if gameStackName == "" || playerStackName == "" {
		//ValidConfiguration reports the reason as a boot error.
		return specs
	}

	return append(specs, legal.MayMoveFirstTo("game."+gameStackName, "player."+playerStackName))
}

// LegalPlanEnabled always returns true: this move type's legality IS its
// contributed preconditions.
func (d *DrawToPlayer) LegalPlanEnabled() bool {
	return true
}

// Apply moves the first component of the draw stack to the current player's
// stack. An embedding move with bookkeeping of its own should override Apply
// and super-call this first.
func (d *DrawToPlayer) Apply(state boardgame.State) error {

	concrete := d.Info().ConcreteMove()

	gameStacker, ok := concrete.(interfaces.GameStacker)

	if !ok {
		return errors.New("move unexpectedly does not implement GameStacker")
	}

	playerStacker, ok := concrete.(interfaces.PlayerStacker)

	if !ok {
		return errors.New("move unexpectedly does not implement PlayerStacker")
	}

	source := gameStacker.GameStack(state.GameState())

	if source == nil {
		return errors.New("the draw stack was nil")
	}

	currentPlayer := state.CurrentPlayer()

	if currentPlayer == nil {
		return errors.New("the game has no valid current player")
	}

	destination := playerStacker.PlayerStack(currentPlayer)

	if destination == nil {
		return errors.New("the player's stack was nil")
	}

	component := source.First()

	if component == nil {
		return errors.New("the draw stack was empty")
	}

	return component.MoveToFirstSlot(destination)
}

// ValidConfiguration resolves the source and destination once, at boot, so an
// ambiguous or missing stack is a NewGameManager error naming the reason.
func (d *DrawToPlayer) ValidConfiguration(exampleState boardgame.State) error {

	if err := d.CurrentPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}

	plan, err := drawPlanFor(d.Info().ConcreteMove(), exampleState)

	if err != nil {
		return err
	}

	//Record the resolution for ContributedPreconditions, which is called
	//later in the boot sequence and is handed no state of its own.
	config := d.CustomConfiguration()
	config[configPropResolvedDrawSource] = plan.gameStackName
	config[configPropResolvedDrawDestination] = plan.playerStackName

	concrete := d.Info().ConcreteMove()

	gameStacker, ok := concrete.(interfaces.GameStacker)

	if !ok {
		return errors.New("move doesn't implement GameStacker")
	}

	if gameStacker.GameStack(exampleState.GameState()) == nil {
		return errors.New("GameStack returned nil")
	}

	playerStacker, ok := concrete.(interfaces.PlayerStacker)

	if !ok {
		return errors.New("move doesn't implement PlayerStacker")
	}

	if playerStacker.PlayerStack(exampleState.PlayerStates()[0]) == nil {
		return errors.New("PlayerStack returned nil")
	}

	return nil
}

// FallbackName returns "Draw To Player".
func (d *DrawToPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Draw To Player"
}

// FallbackHelpText describes what the move does.
func (d *DrawToPlayer) FallbackHelpText() string {
	return "Draws a component from the game's draw stack into the current player's stack"
}

func (*DrawToPlayer) consumesGamePropertyConfiguration()   {}
func (*DrawToPlayer) consumesPlayerPropertyConfiguration() {}

var _ interfaces.GameStacker = (*DrawToPlayer)(nil)
var _ interfaces.PlayerStacker = (*DrawToPlayer)(nil)
var _ PreconditionsProvider = (*DrawToPlayer)(nil)
