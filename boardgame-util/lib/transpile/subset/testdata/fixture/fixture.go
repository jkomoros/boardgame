// Package fixture holds one small function per subset rule. Each function's
// doc comment says which refusal reasons the checker must return for it
// ("want: none" when it is in the subset). subset_test.go checks every one.
package fixture

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

var written string

var constantNames = []string{"a", "b"}

const limit = 3

var enums = enum.NewSet()

var spacesEnum = enums.MustAddRange("spaces", 3, 3)

var notFramework = strings.Repeat("x", 2)

type token struct {
	X int
}

type shape interface {
	Area() int
}

type moveX struct {
	moves.CurrentPlayer
}

type moveY struct {
	moves.DefaultComponent
}

type moveZ struct {
	moves.StartPhase
}

// teamPlayer embeds a behavior the engine does not port (EngineGaps).
type teamPlayer struct {
	behaviors.PlayerTeam
}

// moveW embeds a stock move the engine cannot build (EngineGaps).
type moveW struct {
	moves.Increment
}

// want: package-var-write
func setWritten() {
	written = "x"
}

// want: package-var-read
func readWritten() string {
	return written
}

// want: package-var-read
func readComputedVar() string {
	return notFramework
}

// want: none
func readConstant() string {
	return constantNames[0] + strings.Repeat("y", limit)
}

// want: none
func readFramework() int {
	return len(spacesEnum.ValueToRange(0))
}

// want: unknown-call
func unknownCall() string {
	return strings.Title("x")
}

// want: none
func swapByKey(s boardgame.Stack) error {
	return s.SwapComponentsByKey(0, 1)
}

// want: engine-gap
func engineGap(m *moveW, state boardgame.State) error {
	return m.Increment.Apply(state)
}

// want: unknown-type
func unknownType(x float32) float32 {
	return x
}

// want: map-range
func mapRange(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// want: none
func mapIndex(m map[string]int) int {
	m["a"]++
	return m["a"]
}

// want: slice-alias
func appendAlias(xs []int) []int {
	ys := append(xs, 1)
	return ys
}

// want: slice-alias
func subsliceWritten(xs []int) []int {
	ys := xs[1:]
	ys[0] = 2
	return ys
}

// want: none
func subsliceRead(xs []int) int {
	ys := xs[1:]
	return len(ys)
}

// want: string-bytes
func stringIndex(s string) bool {
	return s[0] == 'a'
}

// want: none
func stringCompare(a, b string) bool {
	return a < b && len(a) > 0
}

// want: struct-copy
func structCopy() int {
	t := token{X: 1}
	return t.X
}

// want: pointer-local
func pointerLocal() int {
	x := 1
	p := &x
	return *p
}

// want: sized-int
func sizedInt(x int32) int {
	return int(x)
}

// want: fixed-array
func fixedArray() int {
	var a [3]int
	return a[0]
}

// want: interface-dispatch, unknown-type
func interfaceDispatch(s shape) int {
	return s.Area()
}

// want: type-switch
func typeSwitch(x boardgame.Move) bool {
	switch x.(type) {
	case *moveZ:
		return true
	}
	return false
}

// want: closure-escape
func closureEscape() {
	f := func() {}
	f()
}

// want: none
func lessCallback(xs []int) {
	sort.SliceStable(xs, func(i, j int) bool { return xs[i] < xs[j] })
}

// want: generic
func generic[T any](x T) T {
	return x
}

// want: goroutine
func goroutine() {
	go setWritten()
}

// want: defer
func deferStatement() {
	defer setWritten()
}

// want: goto
func gotoStatement() {
	goto end
end:
}

// want: fallthrough
func fallthroughStatement(x int) int {
	switch x {
	case 1:
		fallthrough
	default:
		return x
	}
}

// want: reflection
func reflection(x int) bool {
	return reflect.ValueOf(x).IsValid()
}

// want: unsafe
func unsafePointer(p unsafe.Pointer) bool {
	return p != nil
}

// want: foreign-rand
func foreignRand() int {
	return rand.Intn(3)
}

// want: none
func stateRand(state boardgame.State) int {
	return state.Rand().Intn(3)
}

// want: io
func wallClock() {
	time.Sleep(0)
}

// want: sub-ms-duration
func subMillisecond(t boardgame.Timer, m boardgame.Move) {
	t.Start(1500*time.Microsecond, m)
}

// want: none
func wholeMilliseconds(t boardgame.Timer, m boardgame.Move) {
	t.Start(2*time.Second, m)
}

// want: unknown-call
func valueFromString(e enum.Enum) enum.EnumKey {
	return e.ValueFromString("x")
}

// want: unknown-call
func chainAlone(state boardgame.ImmutableState) bool {
	return state.Game() != nil
}

// want: none
func chain(state boardgame.State) boardgame.Move {
	return state.Game().MoveByNameForState("x", state)
}

// want: unknown-call
func badVerb(x int) string {
	return fmt.Sprintf("%x", x)
}

// want: none
func goodVerbs(x int, s string) error {
	return fmt.Errorf("%d %s %v %q", x, s, x, s)
}

// want: none
func errorsAndConversions(p boardgame.PlayerIndex, k enum.EnumKey) error {
	if int(p) > k.Int() {
		return errors.New("bigger")
	}
	return nil
}

// want: none
func (m *moveX) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	return nil
}

// want: none
func (m *moveZ) Apply(state boardgame.State) error {
	return m.StartPhase.Apply(state)
}

// want: unknown-call
func (m *moveX) DefaultsForState(state boardgame.ImmutableState) {
	x := 1
	m.CurrentPlayer.DefaultsForState(state)
	_ = x
}

// want: none
func (m *moveY) Apply(state boardgame.State) error {
	if m.ComponentIndex < 0 {
		return errors.New("negative")
	}
	return nil
}

// want: engine-gap
func readsUnportedBehavior(p *teamPlayer) bool {
	return p.Team == nil
}
