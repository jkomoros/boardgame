package behaviors

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
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

func TestScoreBehavior(t *testing.T) {
	s := &ScoreBehavior{Score: 42}
	assert.For(t, "GameScore returns field value").ThatActual(s.GameScore()).Equals(42)

	s.Score = 0
	assert.For(t, "GameScore returns zero").ThatActual(s.GameScore()).Equals(0)

	s.Score = -5
	assert.For(t, "GameScore returns negative").ThatActual(s.GameScore()).Equals(-5)
}

func TestPlayerElimination(t *testing.T) {
	var b interface{}
	b = &PlayerElimination{}
	_, ok := b.(interfaces.PlayerEliminator)
	assert.For(t, "satisfies PlayerEliminator").ThatActual(ok).IsTrue()

	p := &PlayerElimination{}
	assert.For(t, "initially not eliminated").ThatActual(p.IsEliminated()).IsFalse()

	p.SetEliminated()
	assert.For(t, "eliminated after SetEliminated").ThatActual(p.IsEliminated()).IsTrue()

	p.ClearEliminated()
	assert.For(t, "not eliminated after ClearEliminated").ThatActual(p.IsEliminated()).IsFalse()
}

func TestMoveBudget(t *testing.T) {
	var b interface{}
	b = &MoveBudget{}
	_, ok := b.(interfaces.TurnBudgeter)
	assert.For(t, "satisfies TurnBudgeter").ThatActual(ok).IsTrue()

	m := &MoveBudget{}
	assert.For(t, "initially no moves").ThatActual(m.HasMovesLeft()).IsFalse()

	m.ResetMovesTo(3)
	assert.For(t, "has moves after reset").ThatActual(m.HasMovesLeft()).IsTrue()
	assert.For(t, "3 moves left").ThatActual(m.MovesLeft).Equals(3)

	m.ConsumeMove()
	assert.For(t, "2 moves left").ThatActual(m.MovesLeft).Equals(2)
	assert.For(t, "still has moves").ThatActual(m.HasMovesLeft()).IsTrue()

	m.ConsumeMove()
	m.ConsumeMove()
	assert.For(t, "0 moves left").ThatActual(m.MovesLeft).Equals(0)
	assert.For(t, "no moves left").ThatActual(m.HasMovesLeft()).IsFalse()

	// Bool pattern: budget of 1
	m.ResetMovesTo(1)
	assert.For(t, "has move (bool pattern)").ThatActual(m.HasMovesLeft()).IsTrue()
	m.ConsumeMove()
	assert.For(t, "no move (bool pattern)").ThatActual(m.HasMovesLeft()).IsFalse()
}

func TestDrawDiscardPair(t *testing.T) {
	var b interface{}
	b = &DrawDiscardPair{}
	_, ok := b.(Connectable)
	assert.For(t, "satisfies Connectable").ThatActual(ok).IsTrue()

	_, ok = b.(boardgame.TagConfigurable)
	assert.For(t, "satisfies TagConfigurable").ThatActual(ok).IsTrue()

	// ValidConfiguration should fail before connection
	d := &DrawDiscardPair{}
	err := d.ValidConfiguration(nil)
	assert.For(t, "fails before connect").ThatActual(err).IsNotNil()

	// NeedsReshuffle returns false when not connected
	assert.For(t, "no reshuffle needed when not connected").ThatActual(d.NeedsReshuffle()).IsFalse()

	// Note: more substantive testing with real stacks requires the moves
	// package's test game infrastructure.
}

func TestPlayerTeam(t *testing.T) {
	var b interface{}
	b = &PlayerTeam{}
	_, ok := b.(Connectable)
	assert.For(t, "satisfies Connectable").ThatActual(ok).IsTrue()

	_, ok = b.(interfaces.TeamMember)
	assert.For(t, "satisfies TeamMember").ThatActual(ok).IsTrue()

	// ValidConfiguration should fail before connection
	p := &PlayerTeam{}
	err := p.ValidConfiguration(nil)
	assert.For(t, "fails before connect").ThatActual(err).IsNotNil()

	// TeamMembers returns nil when not connected
	assert.For(t, "nil team members when not connected").ThatActual(p.TeamMembers() == nil).IsTrue()

	// Opponents returns nil when not connected
	assert.For(t, "nil opponents when not connected").ThatActual(p.Opponents() == nil).IsTrue()
}

func TestFaceUpMarket(t *testing.T) {
	var b interface{}
	b = &FaceUpMarket{}
	_, ok := b.(Connectable)
	assert.For(t, "satisfies Connectable").ThatActual(ok).IsTrue()

	_, ok = b.(boardgame.TagConfigurable)
	assert.For(t, "satisfies TagConfigurable").ThatActual(ok).IsTrue()

	// ValidConfiguration should fail before connection
	m := &FaceUpMarket{}
	err := m.ValidConfiguration(nil)
	assert.For(t, "fails before connect").ThatActual(err).IsNotNil()

	// NeedsReplenish returns false when not connected
	assert.For(t, "no replenish needed when not connected").ThatActual(m.NeedsReplenish()).IsFalse()

	// SetDisplaySize works
	m.SetDisplaySize(5)
	assert.For(t, "display size set").ThatActual(m.DisplaySize()).Equals(5)
}

func TestConfigureFromTagsSizeValidation(t *testing.T) {
	m := &FaceUpMarket{}

	// Non-numeric size tag should error
	err := m.ConfigureFromTags(reflect.StructTag(`size:"abc"`), nil)
	assert.For(t, "non-numeric size").ThatActual(err).IsNotNil()

	// Negative size tag should error
	err = m.ConfigureFromTags(reflect.StructTag(`size:"-1"`), nil)
	assert.For(t, "negative size").ThatActual(err).IsNotNil()

	// Explicit zero is ambiguous with an omitted tag and should fail loudly.
	err = m.ConfigureFromTags(reflect.StructTag(`size:"0"`), nil)
	assert.For(t, "zero size").ThatActual(err).IsNotNil()

	// Valid size tag should work
	err = m.ConfigureFromTags(reflect.StructTag(`size:"5"`), nil)
	assert.For(t, "valid size").ThatActual(err).IsNil()
	assert.For(t, "size value").ThatActual(m.DisplaySize()).Equals(5)
}

func TestLocationBehaviorTagConfigurable(t *testing.T) {
	var b interface{}
	b = &LocationBehavior{}
	_, ok := b.(boardgame.TagConfigurable)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestConfigureFromTagsEmptyTag(t *testing.T) {
	// When no location tag is present, ConfigureFromTags should be a no-op.
	l := &LocationBehavior{}
	err := l.ConfigureFromTags(reflect.StructTag(""), nil)
	assert.For(t, "empty tag").ThatActual(err).IsNil()

	err = l.ConfigureFromTags(reflect.StructTag(`unrelated:"foo"`), nil)
	assert.For(t, "unrelated tag").ThatActual(err).IsNil()

	// locationStack should still be nil since no tag was processed.
	assert.For(t, "still needs manual config").ThatActual(l.ValidConfiguration(nil)).IsNotNil()
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
