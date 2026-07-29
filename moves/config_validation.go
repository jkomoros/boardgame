package moves

import (
	"sort"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

type consumesStartPhaseConfiguration interface{ consumesStartPhaseConfiguration() }
type consumesTargetCountConfiguration interface{ consumesTargetCountConfiguration() }
type consumesNumRoundsConfiguration interface{ consumesNumRoundsConfiguration() }
type consumesGamePropertyConfiguration interface{ consumesGamePropertyConfiguration() }
type consumesPlayerPropertyConfiguration interface{ consumesPlayerPropertyConfiguration() }
type consumesLegalTypeConfiguration interface{ consumesLegalTypeConfiguration() }
type consumesAmountConfiguration interface{ consumesAmountConfiguration() }
type consumesExplicitStartConfiguration interface{ consumesExplicitStartConfiguration() }
type consumesUniquenessConfiguration interface{ consumesUniquenessConfiguration() }
type consumesRequireAdminConfiguration interface{ consumesRequireAdminConfiguration() }
type consumesMarketFieldConfiguration interface{ consumesMarketFieldConfiguration() }
type consumesDrawDiscardPairFieldConfiguration interface{ consumesDrawDiscardPairFieldConfiguration() }
type consumesRecordedChoiceConfiguration interface{ consumesRecordedChoiceConfiguration() }

func (*StartPhase) consumesStartPhaseConfiguration() {}

func (*ApplyUntilCount) consumesTargetCountConfiguration()        {}
func (*DealCountComponents) consumesTargetCountConfiguration()    {}
func (*CloseAllSeats) consumesTargetCountConfiguration()          {}
func (*WaitForEnoughPlayers) consumesTargetCountConfiguration()   {}
func (*RoundRobinNumRounds) consumesNumRoundsConfiguration()      {}
func (*DealCountComponents) consumesGamePropertyConfiguration()   {}
func (*DealCountComponents) consumesPlayerPropertyConfiguration() {}
func (*Increment) consumesGamePropertyConfiguration()             {}
func (*Increment) consumesPlayerPropertyConfiguration()           {}
func (*Increment) consumesAmountConfiguration()                   {}
func (*DefaultComponent) consumesLegalTypeConfiguration()         {}
func (*WaitForEnoughPlayers) consumesExplicitStartConfiguration() {}

func (*SelectTeam) consumesUniquenessConfiguration()  {}
func (*SelectRole) consumesUniquenessConfiguration()  {}
func (*SelectColor) consumesUniquenessConfiguration() {}

func (*SelectTeam) consumesRequireAdminConfiguration()                     {}
func (*SelectRole) consumesRequireAdminConfiguration()                     {}
func (*SelectColor) consumesRequireAdminConfiguration()                    {}
func (*CloseAllSeats) consumesRequireAdminConfiguration()                  {}
func (*ReplenishMarket) consumesMarketFieldConfiguration()                 {}
func (*ShuffleDiscardIntoDraw) consumesDrawDiscardPairFieldConfiguration() {}

type configurationCompatibilityCheck struct {
	key      string
	option   string
	accepted func(boardgame.Move, boardgame.PropertyCollection) bool
}

func implementsConfigurationConsumer[T any](move boardgame.Move) bool {
	_, ok := move.(T)
	return ok
}

func validateCustomConfiguration(move boardgame.Move, config boardgame.PropertyCollection) error {
	checks := []configurationCompatibilityCheck{
		{configPropStartPhase, "WithPhaseToStart", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesStartPhaseConfiguration](move)
		}},
		{configPropTargetCount, "WithTargetCount", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesTargetCountConfiguration](move)
		}},
		{configPropNumRounds, "WithNumRounds", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesNumRoundsConfiguration](move)
		}},
		{configPropGameProperty, "WithGameProperty", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesGamePropertyConfiguration](move)
		}},
		{configPropPlayerProperty, "WithPlayerProperty", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesPlayerPropertyConfiguration](move)
		}},
		{configPropLegalType, "WithLegalType", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesLegalTypeConfiguration](move)
		}},
		{configPropAmount, "WithAmount", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesAmountConfiguration](move)
		}},
		{configPropManualStart, "WithManualStart", func(boardgame.Move, boardgame.PropertyCollection) bool { return false }},
		{configPropRequireExplicitStart, "WithRequireExplicitStart", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesExplicitStartConfiguration](move)
		}},
		{configPropUnique, "WithUnique", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesUniquenessConfiguration](move)
		}},
		{configPropAllowDuplicates, "WithAllowDuplicates", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesUniquenessConfiguration](move)
		}},
		{configPropRequireAdmin, "WithRequireAdmin", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesRequireAdminConfiguration](move)
		}},
		{configPropMarketField, "WithMarketField", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesMarketFieldConfiguration](move)
		}},
		{configPropDrawDiscardPairField, "WithDrawDiscardPairField", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesDrawDiscardPairFieldConfiguration](move)
		}},
		{configPropRecordedChoices, "WithRecordedChoice", func(move boardgame.Move, _ boardgame.PropertyCollection) bool {
			return implementsConfigurationConsumer[consumesRecordedChoiceConfiguration](move)
		}},
		//WithSourceProperty and WithDestinationProperty have exactly two
		//consumers: the move's own SourceStack()/DestinationStack(), and
		//moves.Default's generic stack-constraints check -- which reads the
		//PAIR and does nothing with one alone. So one of these passed to a move
		//with no matching stacker method and no partner option is inert, which
		//is precisely the case worth failing at boot. Testing for the interface
		//rather than a marker on the framework base types deliberately also
		//accepts a game's own move that implements the stacker itself.
		{configPropSourceProperty, "WithSourceProperty", func(move boardgame.Move, config boardgame.PropertyCollection) bool {
			if _, ok := move.(interfaces.SourceStacker); ok {
				return true
			}
			return config[configPropDestinationProperty] != nil
		}},
		{configPropDestinationProperty, "WithDestinationProperty", func(move boardgame.Move, config boardgame.PropertyCollection) bool {
			if _, ok := move.(interfaces.DestinationStacker); ok {
				return true
			}
			return config[configPropSourceProperty] != nil
		}},
	}

	var ignored []string
	for _, check := range checks {
		if config[check.key] == nil || check.accepted(move, config) {
			continue
		}
		ignored = append(ignored, check.option)
	}
	if len(ignored) == 0 {
		return nil
	}
	sort.Strings(ignored)
	return errors.New("configuration option has no effect on this move type: " + strings.Join(ignored, ", "))
}
