package boardgame

import (
	"strings"
	"testing"

	stderrors "errors"

	"github.com/sirupsen/logrus"
	"github.com/workfit/tester/assert"
)

/*
legal_logging_test.go exercises #65 (design spec §6, "Server"): fixup-move
rejections are logged at debug level, naming the move, the first-failing
predicate for a move that opted into declarative legality, and the rendered
message — or, for an opaque (non-opted-in) move, just its plain Legal()
error string.
*/

// registerLegalLoggingTestAlwaysFailPredicate registers a trivial
// always-Fail predicate (and its template), reused by
// TestLogFixupRejectionDeclarativeMove below to exercise the "predicate
// name extracted from the plan" branch of legalFixupRejectionPredicateName.
func registerLegalLoggingTestAlwaysFailPredicate() {
	RegisterDefaultLegalPredicateConstructors(&LegalPredicateConstructor{
		Name: "legalLoggingTestAlwaysFail",
		Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
			return &LegalPredicate{
				Name:             "legalLoggingTestAlwaysFail",
				EmittedTemplates: []string{"legalloggingtest.always_fail"},
				Evaluate: func(ctx LegalContext) LegalVerdict {
					return LegalVerdict{
						Outcome: LegalFail,
						Message: &LegalMessage{Template: "legalloggingtest.always_fail"},
					}
				},
			}, nil
		},
	})
	RegisterDefaultLegalTemplates(map[string]string{
		"legalloggingtest.always_fail": "this move is never legal (test fixture)",
	})
}

// legalLoggingOpaqueRejectMove is a fixup move that is NEVER legal, and
// does not opt in to declarative legality: its rejection should log with an
// empty predicate name and the plain error string.
type legalLoggingOpaqueRejectMove struct {
	baseFixUpMove
}

func (m *legalLoggingOpaqueRejectMove) Reader() PropertyReader { return getDefaultReader(m) }
func (m *legalLoggingOpaqueRejectMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}
func (m *legalLoggingOpaqueRejectMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}
func (m *legalLoggingOpaqueRejectMove) Apply(state State) error { return nil }
func (m *legalLoggingOpaqueRejectMove) Legal(state ImmutableState, proposer PlayerIndex) error {
	return stderrors.New("opaque plain rejection")
}

var legalLoggingOpaqueRejectMoveConfig = NewMoveConfig(
	"Logging Opaque Reject",
	func() Move { return new(legalLoggingOpaqueRejectMove) },
	nil,
)

// legalLoggingDeclarativeRejectMoveConfig is a fixup move opted in to
// declarative legality via a single, always-failing top-level spec (reuses
// legal_index_test.go's legalIndexDeclarerMove — same package, and the same
// hand-rolled probe-reaching Legal() contract).
var legalLoggingDeclarativeRejectMoveConfig = NewMoveConfig(
	"Logging Declarative Reject",
	func() Move {
		return &legalLoggingDeclarativeRejectMove{
			legalIndexDeclarerMove: legalIndexDeclarerMove{
				authored: []LegalSpec{{Name: "legalLoggingTestAlwaysFail"}},
			},
		}
	},
	nil,
)

// legalLoggingDeclarativeRejectMove wraps legalIndexDeclarerMove only to
// override IsFixUp (legalIndexDeclarerMove embeds baseMove, which is not a
// fixup move) — the fixup-rejection log path (#65) only fires for isFixUp
// moves (game.go's applyMove).
type legalLoggingDeclarativeRejectMove struct {
	legalIndexDeclarerMove
}

func (m *legalLoggingDeclarativeRejectMove) IsFixUp() bool { return true }

// legalLoggingDeclarativeRejectNonFixUpMoveConfig is the SAME
// always-failing declarative shape as legalLoggingDeclarativeRejectMoveConfig,
// but NOT a fixup move — used by
// TestLogFixupRejectionNotLoggedForNonFixup to prove the log path is scoped
// to isFixUp moves only.
var legalLoggingDeclarativeRejectNonFixUpMoveConfig = NewMoveConfig(
	"Logging Declarative Reject Non FixUp",
	func() Move {
		return &legalIndexDeclarerMove{authored: []LegalSpec{{Name: "legalLoggingTestAlwaysFail"}}}
	},
	nil,
)

// newLegalLoggingTestGame boots a minimal manager with all fixture move
// types installed, and returns a real, running *Game plus a *strings.Builder
// capturing the manager's debug-level log output.
func newLegalLoggingTestGame(t *testing.T) (*Game, *strings.Builder) {
	t.Helper()
	registerLegalLoggingTestAlwaysFailPredicate()

	moveInstaller := func(manager *GameManager) []MoveConfig {
		return []MoveConfig{
			legalLoggingOpaqueRejectMoveConfig,
			legalLoggingDeclarativeRejectMoveConfig,
			legalLoggingDeclarativeRejectNonFixUpMoveConfig,
		}
	}

	manager, err := NewGameManager(&testGameDelegate{moveInstaller: moveInstaller}, newTestStorageManager())
	assert.For(t).ThatActual(err).IsNil()

	var buf strings.Builder
	manager.Logger().SetOutput(&buf)
	manager.Logger().SetLevel(logrus.DebugLevel)

	game, err := manager.newGameImpl("", "")
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(game.setUp(0, nil, nil)).IsNil()

	return game, &buf
}

// TestLogFixupRejectionOpaqueMove proves an opaque move's fixup rejection
// logs at debug with the move name and the PLAIN error string, and no
// predicate name (it has no plan to introspect).
func TestLogFixupRejectionOpaqueMove(t *testing.T) {
	game, buf := newLegalLoggingTestGame(t)
	buf.Reset() // discard setup's own debug logging

	move := game.MoveByName("Logging Opaque Reject")
	assert.For(t).ThatActual(move).IsNotNil()

	err := game.applyMove(move, AdminPlayerIndex, true, 0, selfInitiatorSentinel)
	assert.For(t).ThatActual(err).IsNotNil()

	logged := buf.String()
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "fixup rejected")).Equals(true)
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "Logging Opaque Reject")).Equals(true)
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "opaque plain rejection")).Equals(true)
	// No predicate to report for an opaque move (logrus's TextFormatter
	// renders an empty string value unquoted, as "predicate=" with nothing
	// after it before the line ends).
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "predicate=\n")).Equals(true)
}

// TestLogFixupRejectionDeclarativeMove proves a declaratively-opted-in
// move's fixup rejection logs the FIRST-FAILING PREDICATE's name (design
// spec §6), plus the rendered message.
func TestLogFixupRejectionDeclarativeMove(t *testing.T) {
	game, buf := newLegalLoggingTestGame(t)
	buf.Reset() // discard setup's own debug logging

	move := game.MoveByName("Logging Declarative Reject")
	assert.For(t).ThatActual(move).IsNotNil()

	err := game.applyMove(move, AdminPlayerIndex, true, 0, selfInitiatorSentinel)
	assert.For(t).ThatActual(err).IsNotNil()

	logged := buf.String()
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "fixup rejected")).Equals(true)
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "Logging Declarative Reject")).Equals(true)
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "predicate=legalLoggingTestAlwaysFail")).Equals(true)
	assert.For(t, "log output").ThatActual(strings.Contains(logged, "this move is never legal (test fixture)")).Equals(true)
}

// TestLogFixupRejectionSkippedBelowDebugLevel proves the log-level gate:
// with the logger below Debug, nothing is logged at all (and, implicitly,
// the expensive full-ledger re-evaluation in legalFixupRejectionPredicateName
// never runs).
func TestLogFixupRejectionSkippedBelowDebugLevel(t *testing.T) {
	game, buf := newLegalLoggingTestGame(t)
	game.manager.Logger().SetLevel(logrus.InfoLevel)
	// Setup itself runs a debug-level ProposeFixUpMove pass (base's own
	// logging); reset AFTER setup so this test only observes what the move
	// under test logs.
	buf.Reset()

	move := game.MoveByName("Logging Opaque Reject")
	assert.For(t).ThatActual(move).IsNotNil()

	err := game.applyMove(move, AdminPlayerIndex, true, 0, selfInitiatorSentinel)
	assert.For(t).ThatActual(err).IsNotNil()

	assert.For(t, "log output").ThatActual(buf.Len()).Equals(0)
}

// TestLogFixupRejectionNotLoggedForNonFixup proves #65's scope: a REJECTED
// non-fixup (player) move does NOT go through the fixup-rejection log path
// at all — only isFixUp moves do (game.go's applyMove).
func TestLogFixupRejectionNotLoggedForNonFixup(t *testing.T) {
	game, buf := newLegalLoggingTestGame(t)
	// See TestLogFixupRejectionSkippedBelowDebugLevel: reset after setup's
	// own debug logging.
	buf.Reset()

	move := game.MoveByName("Logging Declarative Reject Non FixUp")
	assert.For(t).ThatActual(move).IsNotNil()

	err := game.applyMove(move, AdminPlayerIndex, false, 0, selfInitiatorSentinel)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t, "log output").ThatActual(buf.Len()).Equals(0)
}
