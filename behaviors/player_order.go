package behaviors

import (
	"errors"
	"fmt"

	"github.com/jkomoros/boardgame"
)

/*
PlayerOrderBehavior is a struct that's designed to be anonymously embedded in
your gameState. It allows you to define a custom player order that will be
used by [boardgame.PlayerIndex.Next] and [boardgame.PlayerIndex.Previous]
instead of the default sequential order. It is a [Connectable] behavior that is
automatically connected by the framework during state setup.

Example:

	type gameState struct {
	    base.SubState
	    behaviors.PlayerOrderBehavior
	    behaviors.CurrentPlayerBehavior
	}
*/
type PlayerOrderBehavior struct {
	container boardgame.SubState
	// OrderSlice is the underlying storage for the custom player order.
	// Exported for codegen serialization only. Do NOT modify directly; use
	// SetPlayerOrder instead, which validates the permutation.
	OrderSlice  []int
	cachedOrder []boardgame.PlayerIndex
	cacheBuilt  bool
}

// ConnectBehavior stores a reference to the container.
func (p *PlayerOrderBehavior) ConnectBehavior(containingSubState boardgame.SubState) {
	p.container = containingSubState
}

// ValidConfiguration returns an error if ConnectBehavior hasn't yet been called.
func (p *PlayerOrderBehavior) ValidConfiguration(example boardgame.State) error {
	if p.container == nil {
		return errors.New("ConnectBehavior hasn't been called. See the behaviors package doc for more on initializing Connectable behaviors")
	}
	return nil
}

// PlayerOrder returns the custom player order, or nil if not set (meaning
// default sequential). This satisfies moves/interfaces.PlayerOrderer.
// The result is validated on first call and cached; subsequent calls return
// a copy of the cached result until SetPlayerOrder is called.
func (p *PlayerOrderBehavior) PlayerOrder() []boardgame.PlayerIndex {
	if len(p.OrderSlice) == 0 {
		return nil
	}

	if !p.cacheBuilt {
		p.cachedOrder = p.buildAndValidateOrder()
		p.cacheBuilt = true
	}

	if p.cachedOrder == nil {
		return nil
	}

	// Return a copy so callers can't corrupt the cache.
	result := make([]boardgame.PlayerIndex, len(p.cachedOrder))
	copy(result, p.cachedOrder)
	return result
}

// buildAndValidateOrder converts OrderSlice to []PlayerIndex and validates
// that values are in-range and unique. Returns nil (meaning default sequential
// order) if validation fails.
func (p *PlayerOrderBehavior) buildAndValidateOrder() []boardgame.PlayerIndex {
	if p.container == nil {
		// Not connected yet; convert without validation.
		result := make([]boardgame.PlayerIndex, len(p.OrderSlice))
		for i, v := range p.OrderSlice {
			result[i] = boardgame.PlayerIndex(v)
		}
		return result
	}

	numPlayers := len(p.container.State().ImmutablePlayerStates())
	logger := p.container.State().Manager().Logger()
	if len(p.OrderSlice) != numPlayers {
		logger.Warnf("PlayerOrderBehavior.OrderSlice length %d does not match number of players %d; using default order", len(p.OrderSlice), numPlayers)
		return nil
	}

	seen := make(map[int]bool, numPlayers)
	result := make([]boardgame.PlayerIndex, len(p.OrderSlice))
	for i, v := range p.OrderSlice {
		if v < 0 || v >= numPlayers {
			logger.Warnf("PlayerOrderBehavior.OrderSlice contains out-of-range index %d; using default order", v)
			return nil
		}
		if seen[v] {
			logger.Warnf("PlayerOrderBehavior.OrderSlice contains duplicate index %d; using default order", v)
			return nil
		}
		seen[v] = true
		result[i] = boardgame.PlayerIndex(v)
	}
	return result
}

// SetPlayerOrder sets a custom order. Must be a valid permutation of
// 0..numPlayers-1 (every index exactly once). Returns error if invalid.
func (p *PlayerOrderBehavior) SetPlayerOrder(order []boardgame.PlayerIndex) error {
	if p.container == nil {
		return errors.New("behavior not connected")
	}
	numPlayers := len(p.container.State().ImmutablePlayerStates())
	if len(order) != numPlayers {
		return fmt.Errorf("order length %d does not match number of players %d", len(order), numPlayers)
	}
	seen := make(map[boardgame.PlayerIndex]bool, numPlayers)
	for _, idx := range order {
		if int(idx) < 0 || int(idx) >= numPlayers {
			return fmt.Errorf("player index %d out of range [0, %d)", idx, numPlayers)
		}
		if seen[idx] {
			return fmt.Errorf("duplicate player index %d", idx)
		}
		seen[idx] = true
	}
	p.OrderSlice = make([]int, len(order))
	for i, idx := range order {
		p.OrderSlice[i] = int(idx)
	}
	// Invalidate cache
	p.cachedOrder = nil
	p.cacheBuilt = false
	return nil
}

// ReversePlayerOrder is a convenience that sets reverse-sequential order.
// Returns the error from SetPlayerOrder (should be nil for valid state).
func (p *PlayerOrderBehavior) ReversePlayerOrder(state boardgame.ImmutableState) error {
	n := len(state.ImmutablePlayerStates())
	order := make([]boardgame.PlayerIndex, n)
	for i := 0; i < n; i++ {
		order[i] = boardgame.PlayerIndex(n - 1 - i)
	}
	return p.SetPlayerOrder(order)
}
