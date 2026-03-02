package behaviors

import (
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"

	"github.com/workfit/tester/assert"
)

func TestRoundRobin(t *testing.T) {
	var b interface{}
	b = &RoundRobin{}
	_, ok := b.(interfaces.RoundRobinProperties)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestCurrentPlayer(t *testing.T) {
	var b interface{}
	b = &CurrentPlayerBehavior{}
	_, ok := b.(interfaces.CurrentPlayerSetter)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestPhase(t *testing.T) {
	var b interface{}
	b = &PhaseBehavior{}
	_, ok := b.(interfaces.CurrentPhaseSetter)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestColor(t *testing.T) {
	var b interface{}
	b = &PlayerColor{}
	_, ok := b.(Connectable)
	assert.For(t).ThatActual(ok).IsTrue()

	//Note that more substantive testing of PlayerColor is done in
	//moves/game_test.go, since testing it requires a whole test game and the
	//moves package has one.
}

func TestInactivePlayer(t *testing.T) {
	var b interface{}
	b = &InactivePlayer{}
	_, ok := b.(interfaces.PlayerInactiver)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestSeater(t *testing.T) {
	var b interface{}
	b = &Seat{}
	_, ok := b.(interfaces.Seater)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestPlayerOrderBehavior(t *testing.T) {
	// Interface satisfaction: PlayerOrderer
	var b interface{}
	b = &PlayerOrderBehavior{}
	_, ok := b.(interfaces.PlayerOrderer)
	assert.For(t).ThatActual(ok).IsTrue()

	// Interface satisfaction: Connectable
	_, ok = b.(Connectable)
	assert.For(t).ThatActual(ok).IsTrue()

	// ValidConfiguration fails before connection
	p := &PlayerOrderBehavior{}
	err := p.ValidConfiguration(nil)
	assert.For(t).ThatActual(err).IsNotNil()

	// PlayerOrder returns nil when empty
	result := p.PlayerOrder()
	assert.For(t).ThatActual(result == nil).IsTrue()

	// SetPlayerOrder fails when not connected
	err = p.SetPlayerOrder(nil)
	assert.For(t).ThatActual(err).IsNotNil()

	// ReversePlayerOrder also fails when not connected (requires container)
	// Note: can't fully test Set/Reverse/PlayerOrder validation without a
	// real game state, which requires the moves package.
}

func TestDefaultPlayerColor(t *testing.T) {
	// Basic colors
	assert.For(t, "red").ThatActual(DefaultPlayerColor(0)).Equals("#D32F2F")
	assert.For(t, "blue").ThatActual(DefaultPlayerColor(1)).Equals("#1976D2")
	assert.For(t, "green").ThatActual(DefaultPlayerColor(2)).Equals("#388E3C")

	// Negative index gets clamped to 0
	assert.For(t, "negative").ThatActual(DefaultPlayerColor(-1)).Equals("#D32F2F")

	// Cycling past palette length
	assert.For(t, "wrap 12").ThatActual(DefaultPlayerColor(12)).Equals("#D32F2F")
	assert.For(t, "wrap 13").ThatActual(DefaultPlayerColor(13)).Equals("#1976D2")
}

func TestCSSColorForKey(t *testing.T) {
	// Verify all named constants have entries
	keys := []enum.EnumKey{ColorRed, ColorBlue, ColorGreen, ColorYellow, ColorBlack, ColorWhite, ColorOrange, ColorPurple, ColorPink, ColorBrown, ColorCyan, ColorGray}
	for _, key := range keys {
		css, ok := CSSColorForKey[key]
		assert.For(t, "key", key).ThatActual(ok).IsTrue()
		assert.For(t, "key", key).ThatActual(len(css) > 0).IsTrue()
	}

	// Verify a non-existent key returns false
	_, ok := CSSColorForKey[999]
	assert.For(t, "missing key").ThatActual(ok).IsFalse()
}

func TestLocationBehavior(t *testing.T) {
	var b interface{}
	b = &LocationBehavior{}
	_, ok := b.(Connectable)
	assert.For(t).ThatActual(ok).IsTrue()

	// ValidConfiguration should fail before connection
	l := &LocationBehavior{}
	err := l.ValidConfiguration(nil)
	assert.For(t).ThatActual(err).IsNotNil()

	// LocationIndex returns nil when no stack/graph connected
	assert.For(t).ThatActual(l.LocationIndex() == nil).IsTrue()

	// LocationIndexKey returns (0, false) when no stack connected
	key, found := l.LocationIndexKey()
	assert.For(t).ThatActual(key).Equals(enum.EnumKey(0))
	assert.For(t).ThatActual(found).IsFalse()

	// Neighbors returns nil when no graph connected
	assert.For(t).ThatActual(len(l.Neighbors())).Equals(0)

	// IsConnectedTo returns false when no graph connected
	assert.For(t).ThatActual(l.IsConnectedTo(nil)).IsFalse()

	// ShortestPathTo returns error when no graph connected
	_, err = l.ShortestPathTo(nil)
	assert.For(t).ThatActual(err).IsNotNil()

	// DistanceTo returns error when no graph connected
	d, err := l.DistanceTo(nil)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(d).Equals(-1)

	// Graph returns nil when no graph connected
	assert.For(t).ThatActual(l.Graph() == nil).IsTrue()

	// Token returns nil when no stack connected
	assert.For(t).ThatActual(l.Token()).IsNil()

	// Note: more substantive testing of LocationBehavior (with a real game
	// state, SizedStack, and graph) is done in moves/game_test.go since
	// testing those requires a whole test game.
}
