package moves

import (
	"github.com/jkomoros/boardgame"
)

// ChoiceOption configures the finite choice surface installed by WithChoices.
// Implementations are sealed so games cannot smuggle unvalidated candidate
// sources or disclosure policies into the projection protocol.
type ChoiceOption interface {
	applyChoiceOption(*choiceOptions)
}

type choiceOption func(*choiceOptions)

func (o choiceOption) applyChoiceOption(options *choiceOptions) { o(options) }

type choiceOptions struct {
	excludedValues []string
	stackSource    *boardgame.MoveChoiceStackSource
	err            string
}

// ExcludeChoices removes enum sentinels or other values that are never legal
// creator inputs. Exclusions also constrain the canonical proposal domain, so
// a forged proposal cannot submit a value merely because the UI omitted it.
// Player-index choices do not support exclusions.
func ExcludeChoices(values ...string) ChoiceOption {
	copied := append([]string(nil), values...)
	return choiceOption(func(options *choiceOptions) {
		options.excludedValues = append(options.excludedValues, copied...)
	})
}

// FromCurrentPlayerStack enumerates occupied indexes in the named stack on
// the proposing player's state. The projected move field must be an integer.
// Opting in explicitly discloses occupied-slot membership and exact legality
// to that actor for the pinned state snapshot.
func FromCurrentPlayerStack(property string) ChoiceOption {
	return fromStack(boardgame.MoveChoiceStackScopeActorPlayer, property)
}

// FromGameStack enumerates occupied indexes in the named game-state stack.
// The projected move field must be an integer. Use it only when disclosing the
// stack's occupied-slot membership to the actor is safe.
func FromGameStack(property string) ChoiceOption {
	return fromStack(boardgame.MoveChoiceStackScopeGame, property)
}

func fromStack(scope boardgame.MoveChoiceStackScope, property string) ChoiceOption {
	return choiceOption(func(options *choiceOptions) {
		if options.stackSource != nil {
			options.err = "more than one stack source was configured"
			return
		}
		options.stackSource = &boardgame.MoveChoiceStackSource{Scope: scope, Property: property}
	})
}

// WithChoices presents one finite required creator-input field as choices to
// the move's acting player. Player-index fields enumerate the player roster,
// ordinary enum fields enumerate their canonical values, and integer fields
// may opt into an explicit FromCurrentPlayerStack or FromGameStack locator.
// Stack sources enumerate occupied slots only. Each complete binding is
// evaluated by the move's canonical Legal method.
//
// Opting in reveals the set's existence, complete candidate universe, and exact
// availability to the actor. UI copy and layout remain client-owned.
func WithChoices(fieldName string, options ...ChoiceOption) CustomConfigurationOption {
	resolved := choiceOptions{}
	for _, option := range options {
		if option != nil {
			option.applyChoiceOption(&resolved)
		}
	}
	excluded := append([]string(nil), resolved.excludedValues...)
	return func(config boardgame.PropertyCollection) {
		declarations, _ := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)
		var stackSource *boardgame.MoveChoiceStackSource
		if resolved.stackSource != nil {
			copied := *resolved.stackSource
			stackSource = &copied
		}
		config[configPropMoveChoiceProjections] = append(declarations, moveChoiceProjectionDeclaration{
			projection: boardgame.MoveChoiceProjection{
				FieldName:      fieldName,
				StackSource:    stackSource,
				ExcludedValues: append([]string(nil), excluded...),
				Disclosure:     boardgame.MoveChoiceDisclosureActorExact,
			},
			err: resolved.err,
		})
	}
}
