/*

Package examplepkg  is just an example package for testing. Note that some of
the uses of codegen use somewhat odd spacing or capitalization; this is
primarily just to test how resilient the package is to unexpected input.

*/
package examplepkg

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

//This is a normal gameDelegate that should have its ConfigureEnums output,
//because it has ConfigureMoves() but not its own ConfigureEnums.
type gameDelegate struct {
	base.GameDelegate
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return nil
}

//This is a normal gameDelegate that should also have its ConfigureEnums output.
type secondGameDelegate struct {
	base.GameDelegate
}

func (s *secondGameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return nil
}

//This delegate already has a manual configureEnums, so shouldn't have one
//automatically generated.
type alreadyHasEnumsGameDelegate struct {
	base.GameDelegate
}

func (a *alreadyHasEnumsGameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return nil
}

func (a *alreadyHasEnumsGameDelegate) ConfigureEnums() *enum.Set {
	//Because we have htis, we shouldn't export an enums.
	return nil
}

//This delegate shouldn't have ConfigureEnums generated because it has
//AnotherMethodName, not ConfigureMoves.
type fakeGameDelegateWrongMethodName struct {
	base.GameDelegate
}

func (f *fakeGameDelegateWrongMethodName) AnotherMethodName() []boardgame.MoveConfig {
	return nil
}

//This delegate shouldn't have ConfigureEnums generated because the return
//type doesn't match the ConfigureMoves() signature.
type fakeGameDelegateWrongReturnType struct {
	base.GameDelegate
}

func (f *fakeGameDelegateWrongReturnType) ConfigureMoves() *base.GameDelegate {
	return nil
}

//boardgame:codegen
const (
	colorUnknown = iota
	colorBlue
	colorGreen
	colorRed
)

//boardgame:codegen
const (
	phaseUnknown = iota
	phaseMultiWord
	phaseVeryLongName
)

//boardgame:codegen
const (
	//display:""
	fooOverrideBlank = iota
	//This shouldn't have any weird output
	fooBlue
	//This should work even though enum instruction is in second line.
	//display:"Green"
	fooOverride
	//display:"My name is \"Blue\""
	fooOverrideQuoted
)

//boardgame:codegen
const (
	transformExampleNormalTransform = iota
	//transform: lower
	transformExampleLowerCase
	//Transform:UPPER
	transformExampleUpperCase
	//transform:none
	transformExampleNormalConfiguredTransform
)

//boardgame:codegen
//transform: upper
const (
	defaultTransformBlue = iota
	defaultTransformGreen
	//transform:none
	defaultTransformRed
)

const (
	dontIncludeBlue = iota
	dontIncludeGreen
)

//boardgame:codegen
const (
	tree = iota
	treeBlue
	treeGreen
	treeRed
)

//boardgame:codegen
const (
	blam = iota
	blamOne
	blamTwo
	blamThree
	blamOne010One
	blamOne010Two
	blamTwo010One
)

//boardgame:codegen
const (
	PublicGreen = iota
	PublicBlue
)

//boardgame:codegen
const (
	example = iota
	exampleOne
	exampleTwo
	//display:"One > One"
	exampleOneOne
	exampleOne010Two
)

//boardgame:codegen
const (
	multiWordTree = iota
	multiWordTreeBlueGreen
	//MultiWordTreeBlueGreenOne is implied; a consistent int stand-in will be
	//created.
	multiWordTreeBlueGreenOneA
	multiWordTreeBlueGreenOneB
	//The next item will result in a single child named "Two A"
	multiWordTreeBlueGreenTwoA
	//The next item will result in a child of Three followed by a child of
	//A since there's an explicit tree break.
	multiWordTreeBlueGreenThree010A
)

//boardgame:codegen
const (
	skipNode = iota
	//SkipNodeRed is implied but not listed
	skipNodeRed010Circle
)

//boardgame:codegen
const (
	prefixBugWhite = iota
	//This case captures an earlier bug where an extra occurance of the prefix
	//was wiped out for the string value. The desired string value is
	//"Greenprefix Bug"
	prefixBugGreenprefixBug
)

//boardgame:codegen
//combine:"Group"
const (
	blargA = iota
	blargB
)

//boardgame:codegen
//combine:"Group"
const (
	//guarantee no overlap with the combine set
	flargC = blargB + 1 + iota
	flargD
)

//boardgame:codegen all
type myStruct struct {
	base.SubState
	MyInt              int
	MyBool             bool
	MySizedStack       boardgame.Stack
	TheTimer           boardgame.Timer
	EnumVar            enum.Val
	MyIntSlice         []int
	MyBoolSlice        []bool
	MyStringSlice      []string
	MyPlayerIndexSlice []boardgame.PlayerIndex
}

//boardgame:codegen
type roundRobinStruct struct {
	behaviors.RoundRobin
	base.SubState
	MyBool bool
}

/*

Long comment

boardgame:codegen
*/
type structWithManyKeys struct {
	base.SubState
	A int
	B int
	D int
	C int
	E int
	F int
	G int
	H int
	I int
}

//boardgame:codegen
type embeddedStruct struct {
	moves.CurrentPlayer
	MyInt int
}

//boardgame:codegen
type doubleEmbeddedStruct struct {
	embeddedStruct
}

//	 boardgame:codegen
type myOtherStruct struct {
	blarg           int
	MyGrowableStack boardgame.Stack
	ThePlayerIndex  boardgame.PlayerIndex
}

type noReaderStruct struct {
	MyInt int
}

// boardgame:codegen reader
type onlyReader struct {
	MyString string
}

//boardgame:codegen
type includesImmutable struct {
	//The immutable variants are allowed; their Mutable*Prop methods will
	//simply return ErrPropertyImmutable.
	MyStack          boardgame.ImmutableStack
	MyMutableStack   boardgame.Stack
	MyImmutableTimer boardgame.ImmutableTimer
	MyTimer          boardgame.Timer
	MyEnum           enum.ImmutableVal
	MyMutableEnum    enum.Val
}

// boardgame:codegen    readSetter
type upToReadSetter struct {
	MyInt int
}

//boardgame:codegen
type sizedStackExample struct {
	MySizedStack        boardgame.ImmutableSizedStack
	MyMutableSizedStack boardgame.SizedStack
}

//boardgame:codegen
type mergedStackExample struct {
	MyMergedStack boardgame.MergedStack
}

//boardgame:codegen
type rangeValExample struct {
	MyMutableRangeVal enum.RangeVal
	MyRangeVal        enum.ImmutableRangeVal
}

//boardgame:codegen
type treeValExample struct {
	MyTreeVal          enum.TreeVal
	MyImmutableTreeVal enum.ImmutableTreeVal
}
