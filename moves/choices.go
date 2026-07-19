package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
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
}

// ExcludeChoices removes enum sentinels or other values that are never legal
// creator inputs. Exclusions also become canonical legal preconditions, so a
// forged proposal cannot submit a value merely because the UI omitted it.
// Player-index choices do not support exclusions.
func ExcludeChoices(values ...string) ChoiceOption {
	copied := append([]string(nil), values...)
	return choiceOption(func(options *choiceOptions) {
		options.excludedValues = append(options.excludedValues, copied...)
	})
}

// WithChoices presents one finite required creator-input field as choices to
// the move's acting player. The field's resolved input codec determines its
// universe: player-index fields enumerate the player roster and ordinary enum
// fields enumerate their canonical values. Each complete binding is evaluated
// by the move's canonical Legal method.
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
		config[configPropMoveChoiceProjections] = append(declarations, moveChoiceProjectionDeclaration{
			projection: boardgame.MoveChoiceProjection{
				FieldName:      fieldName,
				ExcludedValues: append([]string(nil), excluded...),
				Disclosure:     boardgame.MoveChoiceDisclosureActorExact,
			},
		})
		if len(excluded) == 0 {
			return
		}
		specs := make([]legal.Spec, 0, len(excluded))
		for _, value := range excluded {
			specs = append(specs, legal.PropNotEquals("move."+fieldName, value))
		}
		WithLegalPreconditions(specs...)(config)
	}
}
