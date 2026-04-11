package moves

import (
	"errors"
	"fmt"
	"reflect"

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
// [interfaces.MarketProvider] interface.
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
	FixUp
}

func (r *ReplenishMarket) market(state boardgame.ImmutableState) (*behaviors.FaceUpMarket, error) {
	config := r.CustomConfiguration()
	fieldName, hasFieldConfig := config[configPropMarketField]

	if hasFieldConfig {
		strFieldName, ok := fieldName.(string)
		if !ok {
			return nil, errors.New("MarketField config is not a string")
		}
		// Look up the named field on the gameState via reflection.
		v := reflect.ValueOf(state.ImmutableGameState()).Elem()
		t := v.Type()
		structField, ok := t.FieldByName(strFieldName)
		if !ok {
			return nil, fmt.Errorf("gameState has no field %q", strFieldName)
		}
		fieldVal := v.FieldByIndex(structField.Index)
		market, ok := fieldVal.Addr().Interface().(*behaviors.FaceUpMarket)
		if !ok {
			return nil, fmt.Errorf("field %q is not a *behaviors.FaceUpMarket", strFieldName)
		}
		return market, nil
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
	if err := r.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	market, err := r.market(state)
	if err != nil {
		return err
	}

	if !market.NeedsReplenish() {
		return errors.New("market display is full or source is empty")
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

	return first.MoveToLastSlot(market.DisplayStack())
}

// ValidConfiguration verifies that the market can be found.
func (r *ReplenishMarket) ValidConfiguration(exampleState boardgame.State) error {
	_, err := r.market(exampleState)
	if err != nil {
		return fmt.Errorf("ReplenishMarket: %w", err)
	}
	return r.FixUp.ValidConfiguration(exampleState)
}

// FallbackName returns a descriptive name for the move.
func (r *ReplenishMarket) FallbackName(m *boardgame.GameManager) string {
	return "Replenish Market"
}

// FallbackHelpText returns a description of what the move does.
func (r *ReplenishMarket) FallbackHelpText() string {
	return "Draws a card from the source deck to fill an empty slot in the face-up display."
}
