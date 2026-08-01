package moves

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

const configPropSlotField = fullyQualifiedPackageName + "SlotField"
const configPropSourceSlotField = fullyQualifiedPackageName + "SourceSlotField"

// WithSlotField names the move property that carries the slot index the player
// chose. Omit it when your move has exactly one int property: that field is
// discovered automatically, which is the common case.
func WithSlotField(fieldName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropSlotField] = fieldName
	}
}

// WithSourceSlotField names the move property that selects WHICH component to
// move out of the source stack. Omit it and the source is the first component
// in the source stack -- "take the next piece off the pile", which is what
// place-a-token moves want. Pass the same name as [WithSlotField] for mirrored
// layouts, where one index means both "the card here" and "the slot there".
func WithSourceSlotField(fieldName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropSourceSlotField] = fieldName
	}
}

/*
MoveComponentToSlot is "play a card": it moves one component out of a source
stack and into the slot of a destination stack that the player chose. It is one
of the two most common verbs in the medium, and before it existed four of the
seven example games hand-rolled it.

Configure the two stacks with [WithSourceProperty] and
[WithDestinationProperty], which accept the full stack path grammar -- so
either end may live on gameState, on the current player, or on the player one
of the move's own fields names:

	auto.MustConfig(
	    new(movePlaceToken),
	    moves.WithSourceProperty("players[move.TargetPlayerIndex].UnusedTokens"),
	    moves.WithDestinationProperty("game.Slots"),
	)

The slot index comes from a property on your move. You do not have to say which
one when there is only one int property to choose from; if there are several,
name it with [WithSlotField] and boot will tell you the candidates. By default
the component moved is the FIRST in the source stack; [WithSourceSlotField]
selects a different one, and naming the same field for both gives the mirrored
layout where one index means "this card" and "that slot".

This move declares no Legal() of its own. Its legality is contributed
declaratively (legal.MayMoveFirstToSlot, or legal.MayMoveToSlot when a source
slot field is configured) on top of [CurrentPlayer]'s phase and proposer
checks, so an embedding move keeps the full declarative surface: it may add its
own [WithLegalPreconditions] and its own LegalCustom residue for the rules the
catalog cannot express.

boardgame:codegen
*/
type MoveComponentToSlot struct {
	CurrentPlayer
}

// slotPlan is everything a MoveComponentToSlot-shaped move needs, resolved
// once. Every failure here is a boot error, never a mid-game surprise.
type slotPlan struct {
	sourcePath      parsedStackPath
	destinationPath parsedStackPath
	slotField       string
	//sourceSlotField is "" when the source is the first component.
	sourceSlotField string
}

func slotPlanFor(move boardgame.Move) (slotPlan, error) {

	infoer, ok := move.(moveInfoer)

	if !ok {
		return slotPlan{}, errors.New("move did not expose its custom configuration")
	}

	source, configured, err := configuredStackPath(infoer, configPropSourceProperty)
	if err != nil {
		return slotPlan{}, fmt.Errorf("WithSourceProperty: %w", err)
	}
	if !configured {
		return slotPlan{}, errors.New("no source stack: pass WithSourceProperty")
	}

	destination, configured, err := configuredStackPath(infoer, configPropDestinationProperty)
	if err != nil {
		return slotPlan{}, fmt.Errorf("WithDestinationProperty: %w", err)
	}
	if !configured {
		return slotPlan{}, errors.New("no destination stack: pass WithDestinationProperty")
	}

	slotField, err := resolveSlotField(move, infoer)
	if err != nil {
		return slotPlan{}, err
	}

	sourceSlotField, err := configuredIntMoveField(move, infoer, configPropSourceSlotField, "WithSourceSlotField")
	if err != nil {
		return slotPlan{}, err
	}

	return slotPlan{
		sourcePath:      source,
		destinationPath: destination,
		slotField:       slotField,
		sourceSlotField: sourceSlotField,
	}, nil
}

// resolveSlotField returns the move property carrying the chosen slot: the one
// named by WithSlotField, or -- when the move has exactly one int property --
// that one, discovered.
func resolveSlotField(move boardgame.Move, infoer moveInfoer) (string, error) {

	named, err := configuredIntMoveField(move, infoer, configPropSlotField, "WithSlotField")

	if err != nil {
		return "", err
	}

	if named != "" {
		return named, nil
	}

	candidates := intMoveFields(move)

	switch len(candidates) {
	case 1:
		return candidates[0], nil
	case 0:
		return "", errors.New("this move has no int property to carry the chosen slot; add one (for example `Slot int`) to your move struct")
	}

	return "", fmt.Errorf("this move has more than one int property (%v), so the one carrying the chosen slot is ambiguous; name it with WithSlotField", candidates)
}

func configuredIntMoveField(move boardgame.Move, infoer moveInfoer, configPropName, optionName string) (string, error) {

	val, ok := infoer.CustomConfiguration()[configPropName]

	if !ok {
		return "", nil
	}

	name, ok := val.(string)

	if !ok || name == "" {
		return "", fmt.Errorf("%s was not given a property name", optionName)
	}

	propType, ok := move.ReadSetter().Props()[name]

	if !ok {
		return "", fmt.Errorf("%s named %q, which is not a property on this move", optionName, name)
	}

	if propType != boardgame.TypeInt {
		return "", fmt.Errorf("%s named %q, which is not an int property", optionName, name)
	}

	return name, nil
}

func intMoveFields(move boardgame.Move) []string {
	var result []string
	for name, propType := range move.ReadSetter().Props() {
		if propType == boardgame.TypeInt {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

// legalPath renders a parsed stack path in the legal package's path grammar,
// which -- unlike this package's configuration spelling -- has no unqualified
// form.
func (p parsedStackPath) legalPath() string {
	switch p.kind {
	case stackPathCurrentPlayer:
		return "player." + p.prop
	case stackPathMoveField:
		return "players[move." + p.moveField + "]." + p.prop
	}
	return "game." + p.prop
}

// definingPrecondition returns the PreconditionName of the one atom that IS
// this move, which depends on whether a source slot field is configured.
func (m *MoveComponentToSlot) definingPrecondition() PreconditionName {
	plan, err := slotPlanFor(m.Info().ConcreteMove())
	if err == nil && plan.sourceSlotField != "" {
		return PreconditionMayMoveToSlot
	}
	return PreconditionMayMoveFirstToSlot
}

// ContributedPreconditions returns CurrentPlayer's phase and proposer checks
// plus the one atom that IS this move: may the selected component go into the
// selected slot.
func (m *MoveComponentToSlot) ContributedPreconditions() []legal.Spec {

	specs := m.CurrentPlayer.ContributedPreconditions()

	//Default derives a generic "stackConstraints" atom from
	//WithSourceProperty+WithDestinationProperty whenever both name gameState
	//stacks. That atom asks whether the FIRST component of the source would be
	//accepted ANYWHERE in the destination -- neither the component nor the
	//slot this move is about. The slot-aware atom appended below is the honest
	//version of the same question, so keeping both would only add a way to
	//reject a move this move's own check (and its own Apply) considers legal,
	//ahead of the proposer check and of anything the game authored. See
	//examples/memory's Reveal Card, where the generic atom would have taken
	//over the endgame message the slot-aware one and the game's own gates own.
	specs = specsWithout(specs, string(PreconditionStackConstraints))

	plan, err := slotPlanFor(m.Info().ConcreteMove())

	if err != nil {
		//ValidConfiguration reports this as a boot error with the reason; there
		//is nothing honest to contribute in the meantime.
		return specs
	}

	source := plan.sourcePath.legalPath()
	destination := plan.destinationPath.legalPath()
	slot := "move." + plan.slotField

	if plan.sourceSlotField != "" {
		return append(specs, legal.MayMoveToSlot(source, destination, "move."+plan.sourceSlotField, slot))
	}

	return append(specs, legal.MayMoveFirstToSlot(source, destination, slot))
}

// specsWithout returns specs minus every spec named name.
func specsWithout(specs []legal.Spec, name string) []legal.Spec {
	result := make([]legal.Spec, 0, len(specs))
	for _, spec := range specs {
		if spec.Name == name {
			continue
		}
		result = append(result, spec)
	}
	return result
}

// LegalPlanEnabled always returns true: this move type's legality IS its
// contributed preconditions, so a plan is assembled whether or not the game
// authored any specs of its own.
func (m *MoveComponentToSlot) LegalPlanEnabled() bool {
	return true
}

// SourceStack returns the stack named by WithSourceProperty.
func (m *MoveComponentToSlot) SourceStack(state boardgame.State) boardgame.Stack {
	return resolveConfiguredStack(m.Info().ConcreteMove(), configPropSourceProperty, state)
}

// DestinationStack returns the stack named by WithDestinationProperty.
func (m *MoveComponentToSlot) DestinationStack(state boardgame.State) boardgame.Stack {
	return resolveConfiguredStack(m.Info().ConcreteMove(), configPropDestinationProperty, state)
}

// Apply moves the selected component into the selected slot. An embedding move
// that has bookkeeping of its own should override Apply and super-call this
// first.
func (m *MoveComponentToSlot) Apply(state boardgame.State) error {

	concrete := m.Info().ConcreteMove()

	plan, err := slotPlanFor(concrete)

	if err != nil {
		return err
	}

	stacker, ok := concrete.(sourceDestinationStacker)

	if !ok {
		return errors.New("move unexpectedly does not have Source/Destination stackers")
	}

	source, destination := stacker.SourceStack(state), stacker.DestinationStack(state)

	if source == nil {
		return errors.New("source stack was nil")
	}

	if destination == nil {
		return errors.New("destination stack was nil")
	}

	slot, err := concrete.ReadSetter().IntProp(plan.slotField)

	if err != nil {
		return fmt.Errorf("couldn't read slot from %q: %w", plan.slotField, err)
	}

	component := source.First()

	if plan.sourceSlotField != "" {
		sourceIndex, err := concrete.ReadSetter().IntProp(plan.sourceSlotField)
		if err != nil {
			return fmt.Errorf("couldn't read source index from %q: %w", plan.sourceSlotField, err)
		}
		component = source.ComponentAt(sourceIndex)
	}

	if component == nil {
		return errors.New("there was no component to move")
	}

	return component.MoveTo(destination, slot)
}

// ValidConfiguration verifies at boot that both stacks resolve, that the slot
// field is unambiguous and int-typed, and that a configured source slot field
// exists.
func (m *MoveComponentToSlot) ValidConfiguration(exampleState boardgame.State) error {

	if err := m.CurrentPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}

	concrete := m.Info().ConcreteMove()

	if err := validateConfiguredStack(concrete, configPropSourceProperty, "WithSourceProperty", exampleState); err != nil {
		return err
	}

	if err := validateConfiguredStack(concrete, configPropDestinationProperty, "WithDestinationProperty", exampleState); err != nil {
		return err
	}

	if _, err := slotPlanFor(concrete); err != nil {
		return err
	}

	if err := requireDefiningPreconditionReauthored(m, "MoveComponentToSlot", m.definingPrecondition(), "legal."+reauthorConstructorFor(m.definingPrecondition())); err != nil {
		return err
	}

	stacker, ok := concrete.(sourceDestinationStacker)

	if !ok {
		return errors.New("move doesn't have Source/Destination stackers")
	}

	if stacker.SourceStack(exampleState) == nil {
		return errors.New("SourceStack returned nil")
	}

	if stacker.DestinationStack(exampleState) == nil {
		return errors.New("DestinationStack returned nil")
	}

	return nil
}

// FallbackName returns a name derived from the two stacks.
func (m *MoveComponentToSlot) FallbackName(g *boardgame.GameManager) string {
	source, destination := m.configuredStackNames()
	return "Move Component From " + source + " To A Slot In " + destination
}

// FallbackHelpText describes what the move does.
func (m *MoveComponentToSlot) FallbackHelpText() string {
	source, destination := m.configuredStackNames()
	return "Moves a component from " + source + " into the chosen slot of " + destination
}

func (m *MoveComponentToSlot) configuredStackNames() (source, destination string) {
	return stackName(m, configPropSourceProperty, nil, nil), stackName(m, configPropDestinationProperty, nil, nil)
}

var _ interfaces.SourceStacker = (*MoveComponentToSlot)(nil)
var _ interfaces.DestinationStacker = (*MoveComponentToSlot)(nil)
var _ PreconditionsProvider = (*MoveComponentToSlot)(nil)
