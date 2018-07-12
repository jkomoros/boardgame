/*

	privateconstants is just the private constant names for CustomConfig
	parameters factored out into a separate package so that multiple sub-
	packages can use the same constants and not have to duplicate them, which
	would be error prone.

	No other package should use these.

*/
package privateconstants

const fullyQualifiedPackageName = "github.com/jkomoros/boardgame/moves."

const StartPhase = fullyQualifiedPackageName + "StartPhase"
const StartPhaseEnum = fullyQualifiedPackageName + "StartPhaseEnum"
const SourceStack = fullyQualifiedPackageName + "SourceStack"
const DestinationStack = fullyQualifiedPackageName + "DestinationStack"
const TargetCount = fullyQualifiedPackageName + "TargetCount"
const NumRounds = fullyQualifiedPackageName + "NumRounds"
const GameStack = fullyQualifiedPackageName + "GameStack"
const PlayerStack = fullyQualifiedPackageName + "PlayerStack"
const MoveName = fullyQualifiedPackageName + "MoveName"
const HelpText = fullyQualifiedPackageName + "HelpText"
const IsFixUp = fullyQualifiedPackageName + "IsFixUp"
const LegalPhases = fullyQualifiedPackageName + "LegalPhases"
const LegalMoveProgression = fullyQualifiedPackageName + "LegalMoveProgression"
const LegalType = fullyQualifiedPackageName + "LegalType"
