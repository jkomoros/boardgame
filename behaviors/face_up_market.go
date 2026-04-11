package behaviors

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jkomoros/boardgame"
)

/*
FaceUpMarket tracks a source deck and a face-up display area on a gameState.
When cards are taken from the display, a companion FixUp move replenishes it
from the source deck. It is a [Connectable], [boardgame.TagConfigurable]
behavior.

Configure it with struct tags pointing to existing Stack fields on the same
struct. The "size" tag specifies how many cards the display should hold:

	type gameState struct {
	    base.SubState
	    behaviors.FaceUpMarket `source:"MerchantCards" display:"CurrentMerchantCards" size:"5"`
	    MerchantCards        boardgame.Stack `stack:"merchants" sanitize:"len"`
	    CurrentMerchantCards boardgame.Stack `stack:"merchants"`
	}

The companion move [moves.ReplenishMarket] works with this behavior to
automatically fill the display when it has fewer than DisplaySize components.

# Multiple Markets

Games with multiple markets (e.g. separate merchant and point card displays)
use named fields. The framework's autoConnectBehaviors processes all struct
fields, not just anonymous ones, so named fields with struct tags are
auto-wired too:

	type gameState struct {
	    base.SubState
	    MerchantMarket behaviors.FaceUpMarket `source:"MerchantCards" display:"CurrentMerchantCards" size:"5"`
	    PointMarket    behaviors.FaceUpMarket `source:"PointCards" display:"CurrentPointCards" size:"5"`
	}

Each market gets its own companion move:

	auto.MustConfig(new(moves.ReplenishMarket), moves.WithMarketField("MerchantMarket"))
	auto.MustConfig(new(moves.ReplenishMarket), moves.WithMarketField("PointMarket"))
*/
type FaceUpMarket struct {
	container    boardgame.SubState
	sourceStack  boardgame.Stack
	displayStack boardgame.Stack
	displaySize  int
}

const sourceStructTag = "source"
const displayStructTag = "display"
const sizeStructTag = "size"

// ConnectBehavior stores a reference to the containing SubState.
func (m *FaceUpMarket) ConnectBehavior(containingSubState boardgame.SubState) {
	m.container = containingSubState
}

// ConfigureFromTags reads the "source", "display", and "size" struct tags from
// the embedding site and configures the behavior.
func (m *FaceUpMarket) ConfigureFromTags(tags reflect.StructTag, containingSubState boardgame.SubState) error {
	if name := tags.Get(sourceStructTag); name != "" {
		stack, err := lookupStackField(containingSubState, name)
		if err != nil {
			return fmt.Errorf("FaceUpMarket: source tag: %w", err)
		}
		m.sourceStack = stack
	}
	if name := tags.Get(displayStructTag); name != "" {
		stack, err := lookupStackField(containingSubState, name)
		if err != nil {
			return fmt.Errorf("FaceUpMarket: display tag: %w", err)
		}
		m.displayStack = stack
	}
	if sizeStr := tags.Get(sizeStructTag); sizeStr != "" {
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("FaceUpMarket: size tag %q is not a valid integer: %w", sizeStr, err)
		}
		if size <= 0 {
			return fmt.Errorf("FaceUpMarket: size tag must be positive, got %d", size)
		}
		m.displaySize = size
	}
	return nil
}

// ConnectSourceStack manually sets the source stack.
func (m *FaceUpMarket) ConnectSourceStack(stack boardgame.Stack) {
	m.sourceStack = stack
}

// ConnectDisplayStack manually sets the display stack.
func (m *FaceUpMarket) ConnectDisplayStack(stack boardgame.Stack) {
	m.displayStack = stack
}

// SetDisplaySize sets the target number of components in the display.
func (m *FaceUpMarket) SetDisplaySize(size int) {
	m.displaySize = size
}

// ValidConfiguration returns an error if the behavior hasn't been properly
// connected.
func (m *FaceUpMarket) ValidConfiguration(example boardgame.State) error {
	if m.container == nil {
		return errors.New("FaceUpMarket: ConnectBehavior hasn't been called. See the behaviors package doc for more on initializing Connectable behaviors")
	}
	if m.sourceStack == nil {
		return errors.New("FaceUpMarket: source stack not connected. Use the `source` struct tag or call ConnectSourceStack in FinishStateSetUp")
	}
	if m.displayStack == nil {
		return errors.New("FaceUpMarket: display stack not connected. Use the `display` struct tag or call ConnectDisplayStack in FinishStateSetUp")
	}
	if m.displaySize <= 0 {
		return errors.New("FaceUpMarket: display size not set. Use the `size` struct tag or call SetDisplaySize in FinishStateSetUp")
	}
	return nil
}

// SourceStack returns the connected source stack.
func (m *FaceUpMarket) SourceStack() boardgame.Stack {
	return m.sourceStack
}

// DisplayStack returns the connected display stack.
func (m *FaceUpMarket) DisplayStack() boardgame.Stack {
	return m.displayStack
}

// DisplaySize returns the target number of components in the display.
func (m *FaceUpMarket) DisplaySize() int {
	return m.displaySize
}

// NeedsReplenish returns true if the display has fewer than DisplaySize
// components and the source stack has components available.
func (m *FaceUpMarket) NeedsReplenish() bool {
	if m.displayStack == nil || m.sourceStack == nil {
		return false
	}
	return m.displayStack.NumComponents() < m.displaySize && m.sourceStack.NumComponents() > 0
}

// GetFaceUpMarket returns itself, satisfying [HasFaceUpMarket].
func (m *FaceUpMarket) GetFaceUpMarket() *FaceUpMarket {
	return m
}

// HasFaceUpMarket is implemented by any SubState that anonymously embeds a
// FaceUpMarket. It allows companion moves to find the behavior via type
// assertion when only one market exists. For multiple markets, use named fields
// and configure companion moves with the field name.
type HasFaceUpMarket interface {
	GetFaceUpMarket() *FaceUpMarket
}
