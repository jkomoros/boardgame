package moves

import (
	"errors"
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

const configPropMarketField = fullyQualifiedPackageName + "MarketField"

// WithMarketField returns a configuration option for [ReplenishMarket] that
// specifies which named FaceUpMarket field on the gameState to target. This is
// required when the gameState has multiple FaceUpMarket fields (as named, not
// anonymous, fields). When the gameState has a single anonymous FaceUpMarket,
// this is not needed -- the move auto-discovers it via [HasFaceUpMarket].
func WithMarketField(fieldName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropMarketField] = fieldName
	}
}

// ReplenishMarket is a FixUp move that automatically draws components from a
// source deck to fill a face-up display area. It works with gameStates that
// embed [behaviors.FaceUpMarket], which satisfies the
// [behaviors.HasFaceUpMarket] interface.
//
// For a single anonymous FaceUpMarket, usage is zero-config:
//
//	auto.MustConfig(new(moves.ReplenishMarket))
//
// For named fields (multiple markets), specify which field:
//
//	auto.MustConfig(new(moves.ReplenishMarket), moves.WithMarketField("MerchantMarket"))
//
//boardgame:codegen
type ReplenishMarket struct {
	FixUpMulti
}

func (r *ReplenishMarket) market(state boardgame.ImmutableState) (*behaviors.FaceUpMarket, error) {
	config := r.CustomConfiguration()
	fieldName, hasFieldConfig, err := configuredString(config, configPropMarketField, "WithMarketField")
	if err != nil {
		return nil, err
	}
	if hasFieldConfig {
		value, err := namedBehaviorField(state, fieldName, new(behaviors.FaceUpMarket))
		if err != nil {
			return nil, err
		}
		return value.(*behaviors.FaceUpMarket), nil
	}

	// Auto-discover via HasFaceUpMarket
	if hasMarket, ok := state.ImmutableGameState().(behaviors.HasFaceUpMarket); ok {
		return hasMarket.GetFaceUpMarket(), nil
	}

	return nil, errors.New("gameState does not implement HasFaceUpMarket and no WithMarketField was configured")
}

// Legal returns nil when the market's display has fewer components than its
// target size and the source has components available.
func (r *ReplenishMarket) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := r.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}

	market, err := r.market(state)
	if err != nil {
		return err
	}

	if !market.NeedsReplenish() {
		return errors.New("market display is full or source is empty")
	}
	first := market.SourceStack().First()
	if first == nil {
		return errors.New("market source stack has no components")
	}
	display := market.DisplayStack()
	slot := display.Len()
	if sized := display.SizedStack(); sized != nil {
		slot = sized.NextSlot()
	}
	if slot < 0 {
		return errors.New("market display has no available slot")
	}
	if err := first.MayMoveToSlot(display, slot); err != nil {
		return fmt.Errorf("market cannot replenish its next slot: %w", err)
	}
	return nil
}

// Apply moves one component from the source stack to the display stack. The
// FixUp will be applied repeatedly until the display reaches its target size.
func (r *ReplenishMarket) Apply(state boardgame.State) error {
	market, err := r.market(state)
	if err != nil {
		return err
	}

	first := market.SourceStack().First()
	if first == nil {
		return errors.New("source stack has no components")
	}

	return first.MoveToNextSlot(market.DisplayStack())
}

// ValidConfiguration verifies that the market can be found.
func (r *ReplenishMarket) ValidConfiguration(exampleState boardgame.State) error {
	market, err := r.market(exampleState)
	if err != nil {
		return fmt.Errorf("ReplenishMarket: %w", err)
	}
	if market == nil {
		return errors.New("ReplenishMarket: face-up market provider returned nil")
	}
	if err := market.ValidConfiguration(exampleState); err != nil {
		return fmt.Errorf("ReplenishMarket: %w", err)
	}
	return r.FixUpMulti.ValidConfiguration(exampleState)
}

// FallbackName returns a descriptive name for the move.
func (r *ReplenishMarket) FallbackName(m *boardgame.GameManager) string {
	if field, configured, err := configuredString(r.CustomConfiguration(), configPropMarketField, "WithMarketField"); configured && err == nil {
		return "Replenish " + titleCaseToWords(field)
	}
	return "Replenish Market"
}

// FallbackHelpText returns a description of what the move does.
func (r *ReplenishMarket) FallbackHelpText() string {
	return "Draws a card from the source deck to fill an empty slot in the face-up display."
}
