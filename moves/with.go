package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
)

const fullyQualifiedPackageName = "github.com/jkomoros/boardgame/moves."

const configPropStartPhase = fullyQualifiedPackageName + "StartPhase"
const configPropStartPhaseEnum = fullyQualifiedPackageName + "StartPhaseEnum"
const configPropSourceProperty = fullyQualifiedPackageName + "SourceProperty"
const configPropDestinationProperty = fullyQualifiedPackageName + "DestinationProperty"
const configPropTargetCount = fullyQualifiedPackageName + "TargetCount"
const configPropNumRounds = fullyQualifiedPackageName + "NumRounds"
const configPropGameProperty = fullyQualifiedPackageName + "GameProperty"
const configPropPlayerProperty = fullyQualifiedPackageName + "PlayerPropety"
const configPropMoveName = fullyQualifiedPackageName + "MoveName"
const configPropMoveNameSuffix = fullyQualifiedPackageName + "MoveNameSuffix"
const configPropHelpText = fullyQualifiedPackageName + "HelpText"
const configPropIsFixUp = fullyQualifiedPackageName + "IsFixUp"
const configPropLegalPhases = fullyQualifiedPackageName + "LegalPhases"
const configPropLegalMoveProgression = fullyQualifiedPackageName + "LegalMoveProgression"
const configPropLegalType = fullyQualifiedPackageName + "LegalType"
const configPropLegalTypeEnum = fullyQualifiedPackageName + "LegalTypeEnum"
const configPropAmount = fullyQualifiedPackageName + "Amount"
const configPropManualStart = fullyQualifiedPackageName + "ManualStart"
const configPropRequireExplicitStart = fullyQualifiedPackageName + "RequireExplicitStart"
const configPropUnique = fullyQualifiedPackageName + "Unique"
const configPropAllowDuplicates = fullyQualifiedPackageName + "AllowDuplicates"
const configPropRequireAdmin = fullyQualifiedPackageName + "RequireAdmin"
const configPropPreconditions = fullyQualifiedPackageName + "Preconditions"
const configPropSuppressedPreconditions = fullyQualifiedPackageName + "SuppressedPreconditions"
const configPropLegalPlanEnabled = fullyQualifiedPackageName + "LegalPlanEnabled"

// CustomConfigurationOption is a function that takes a PropertyCollection and
// modifies a key on it. This package defines a number of functions that return
// funcs that satisfy this interface and can be used in auto.Config to pass in
// configuration to the base moves without requiring verbose embedding and
// method overriding. All of those functions in this package start with "With".
// Config rejects specialized built-in options when the selected move type does
// not embed a base that consumes them, so irrelevant options fail during game
// manager construction instead of being silently ignored.
type CustomConfigurationOption func(boardgame.PropertyCollection)

// WithLegalType returns a function configuration option suitable for being
// passed to auto.Config. The legalType will be passed to the components' Legal()
// method as an ImmutableVal. Idiomatically this should be a value from an enum
// that is related to the legalType for that type of component. The optional
// legalTypeEnum, if provided, will be used to construct the ImmutableVal passed
// to Legal(); if not provided, nil will be passed. If you only have one
// DefaultComponent move for that type of component, it's fine to just skip this.
func WithLegalType(legalType enum.EnumKey, optionalLegalTypeEnum ...enum.Enum) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalType] = legalType
		if len(optionalLegalTypeEnum) > 0 {
			config[configPropLegalTypeEnum] = optionalLegalTypeEnum[0]
		}
	}
}

// WithMoveName returns a function configuration option suitable for being passed
// to auto.Config. moves.Default uses this, if provided, to power MoveTypeName,
// which means that auto.Config will use this name whenever it is passed. If
// you're passing a move struct that not's from this package, the auto-generated
// move name is likely sufficient and you don't need this. See the documentation
// for moves.Default.MoveTypeName for more information.
func WithMoveName(moveName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropMoveName] = moveName
	}
}

// WithMoveNameSuffix returns a function configuration option suitable for being
// passed to auto.Config. The suffix, if provided, will be appended to whatever
// the Move's name would have been (see the behavior for DeriveName on
// move.Default). This is useful because every move must have a unique name, but
// sometimes you have the same underlying move struct who is legal in different
// points in different progressions. This makes it easy to provide a suffix for
// subsequent uses of the same move to ensure the names are all unique.
func WithMoveNameSuffix(suffix string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropMoveNameSuffix] = suffix
	}
}

// WithHelpText returns a function configuration option suitable for being passed
// to auto.Config. moves.Default uses this, if provided, to power
// MoveTypeHelpText, which means that auto.Config will use this text whenever it
// is passed. See the documentation for moves.Default.MoveTypeHelpText for more
// information.
func WithHelpText(helpText string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropHelpText] = helpText
	}
}

// WithLegalPhases returns a function configuration option suitable for being
// passed to auto.Config. legalPhases will extend whatever has already been
// passed before. move.Base will use the result of this to determine if a given
// move is legal in the current phase. Typically you don't use this directly, and
// instead use moves.AddForPhase to use this implicitly.
func WithLegalPhases(legalPhases ...enum.EnumKey) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		previousLegalPhases := config[configPropLegalPhases]

		if keys, ok := previousLegalPhases.([]enum.EnumKey); ok {
			legalPhases = append(keys, legalPhases...)
		}

		config[configPropLegalPhases] = legalPhases
	}
}

// WithLegalMoveProgression returns a function configuration option suitable for
// being passed to auto.Config. moves.Default's Legal() will use this for this
// move type to determine if the move is legal in the order it's being applied.
// Typically you don't use this directly, and instead use
// moves.AddOrderedForPhase to use this implicitly.
func WithLegalMoveProgression(group MoveProgressionGroup) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalMoveProgression] = group
	}
}

// WithIsFixUp returns a function configuration option suitable for being passed
// to auto.Config. moves.Default uses this, if provided, to power
// MoveTypeIsFixUp, which means that auto.Config will use this if it is passed.
// See the documentation for moves.Default.IsFixUp for more information. All
// moves in this package will return reasonable values for IsFixUp on their own,
// so it is much more rare to use this than other config options in this package.
// In general, instead of using this option you should simply embed FixUp (or a
// move that itself embedds IsFixUp), so you don't have to remember to pass
// WithIsFixUp, which is easy to forget.
func WithIsFixUp(isFixUp bool) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropIsFixUp] = isFixUp
	}
}

// WithPhaseToStart returns a function configuration option suitable for being
// passed to auto.Config. PhaseEnum should be the enum that is used for phases,
// and phaseToStart is the value within that phase to start. The phaseEnum is
// optional; if not provided, the name of the move and help text will just use
// the int value of the phase instead.
func WithPhaseToStart(phaseToStart enum.EnumKey, optionalPhaseEnum enum.Enum) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropStartPhase] = phaseToStart
		config[configPropStartPhaseEnum] = optionalPhaseEnum
	}
}

// WithSourceProperty returns a function configuration option suitable for being
// passed to auto.Config. The stackPropName is assumed to be on the GameState
// object. If it isn't, you'll need to embed the move and override SourceStack
// yourself.
func WithSourceProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropSourceProperty] = stackPropName
	}
}

// WithDestinationProperty returns a function configuration option suitable for
// being passed to auto.Config. The stackPropName is assumed to be on the
// GameState object. If it isn't, you'll need to embed the move and override
// DestinationStack yourself.
func WithDestinationProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropDestinationProperty] = stackPropName
	}
}

// WithGameProperty returns a function configuration option suitable for being
// passed to auto.Config. Often used to configure what a move's GameStack() will
// return, but other moves use it for non-stack properties.
func WithGameProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropGameProperty] = stackPropName
	}
}

// WithPlayerProperty returns a function configuration option suitable for being
// passed to auto.Config. Often used to configure what a move's PlayerStack()
// will return, but other moves use it for non-stack properties.
func WithPlayerProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropPlayerProperty] = stackPropName
	}
}

// WithNumRounds returns a function configuration option suitable for being
// passed to auto.Config.
func WithNumRounds(numRounds int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropNumRounds] = numRounds
	}
}

// WithTargetCount returns a function configuration option suitable for being
// passed to auto.Config.
func WithTargetCount(targetCount int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropTargetCount] = targetCount
	}
}

// WithAmount returns a function configuration option suitable for being passed
// to auto.Config.
func WithAmount(amount int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropAmount] = amount
	}
}

// WithManualStart returns a function configuration option suitable for being
// passed to DefaultRoundSetup. When provided, the round setup will include a
// player-callable CloseAllSeats move (displayed as "Confirm Players") and
// configure WaitForEnoughPlayers to block until all unfilled seats are closed.
// This allows players to wait for more people to join before someone explicitly
// starts the game.
func WithManualStart() CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropManualStart] = true
	}
}

// WithRequireExplicitStart returns a function configuration option suitable for
// being passed to auto.Config when configuring WaitForEnoughPlayers. When set,
// WaitForEnoughPlayers will additionally require that all unfilled seats are
// closed before it fires. This is used internally by DefaultRoundSetup when
// WithManualStart is provided, but can also be used directly.
func WithRequireExplicitStart() CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropRequireExplicitStart] = true
	}
}

// WithUnique returns a function configuration option for [SelectTeam],
// [SelectRole], or [SelectColor]. When set, Legal will reject values already
// claimed by another seated player. Useful for games where each player must
// have a unique selection (e.g., Spirit Island spirits). SelectColor already
// defaults to unique, so WithUnique is mainly useful for SelectRole and
// SelectTeam.
func WithUnique() CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropUnique] = true
	}
}

// WithAllowDuplicates returns a function configuration option for [SelectTeam],
// [SelectRole], or [SelectColor]. It disables uniqueness enforcement, allowing
// multiple players to select the same value. SelectTeam and SelectRole already
// default to non-unique, so WithAllowDuplicates is mainly useful for
// SelectColor (which defaults to unique).
func WithAllowDuplicates() CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropAllowDuplicates] = true
	}
}

// WithLegalPreconditions returns a function configuration option suitable for
// being passed to auto.Config. Calling this option opts a move type in to
// declarative legality, even when specs is empty. The specs, in order, are appended
// to whatever has already been passed to earlier WithLegalPreconditions calls
// for this move type (mirroring WithLegalPhases' accumulate-across-calls
// behavior). See moves.Default.DeclaredPreconditions, which reads this
// configuration back out, and moves.PreconditionsProvider /
// moves.Default.ContributedPreconditions for how these authored specs
// combine with the base type's own contributed specs (inPhase/
// inProgression/stackConstraints/proposerIsCurrentPlayer) at plan-assembly
// time. Declaring WithLegalPreconditions does not, by itself, change Legal()'s
// behavior for a move that also overrides Legal() without super-calling
// into the frozen chain — see the design spec's "prime guarantee" for the
// boot-time probe that catches that mistake.
func WithLegalPreconditions(specs ...legal.Spec) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalPlanEnabled] = true
		previous, _ := config[configPropPreconditions].([]legal.Spec)
		config[configPropPreconditions] = append(previous, specs...)
	}
}

// WithoutLegalPrecondition returns a function configuration option suitable for
// being passed to auto.Config. It suppresses one CONTRIBUTED precondition
// (one of the framework's own stable names — pass the exported constants
// [PreconditionInPhase], [PreconditionInProgression],
// [PreconditionStackConstraints], [PreconditionProposerIsCurrentPlayer]
// rather than raw strings — design spec §2) by name, for a move type that
// wants to opt out of an inherited check entirely (the
// moves.ForceFinishTurn "inherit nothing" pattern, now expressible
// declaratively). Suppression names accumulate across multiple calls, like
// WithLegalPreconditions accumulates specs. It does not remove an AUTHORED spec
// passed via WithLegalPreconditions; those are simply not passed in the first
// place.
//
// Suppressions are validated by NewGameManager's boot gauntlet, so a
// WithoutLegalPrecondition call can never silently do nothing: a name that
// matches no spec the move actually contributes (a typo, or a check the
// move never had — e.g. PreconditionInProgression on a move with no move
// progression) is a boot error listing the move's real contributed names.
// Calling WithoutLegalPrecondition opts the move into declarative legality;
// the suppression is itself an explicit request to assemble a plan. Every
// suppression is boot-validated against the move's actual contributions.
func WithoutLegalPrecondition(name PreconditionName) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalPlanEnabled] = true
		previous, _ := config[configPropSuppressedPreconditions].([]string)
		config[configPropSuppressedPreconditions] = append(previous, string(name))
	}
}

// WithRequireAdmin returns a function configuration option for moves like
// [CloseAllSeats]. When set,
// Legal will verify that the proposer is the game administrator (the player
// whose [behaviors.GameAdministrator].IsAdmin is true). If the playerState
// does not embed [behaviors.GameAdministrator], the check is skipped.
// [AdminPlayerIndex] always passes admin checks.
func WithRequireAdmin() CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropRequireAdmin] = true
	}
}
