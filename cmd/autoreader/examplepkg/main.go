/*

	This is just an example package for testing. Note that some of the uses of
	autoreader use somewhat odd spacing or capitalization; this is primarily
	just to test how resilient the package is to unexpected input.

*/
package examplepkg

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
)

//go:generate autoreader

//This is a normal gameDelegate that should have its ConfigureEnums output,
//because it has ConfigureMoves() but not its own ConfigureEnums.
type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return nil
}

//This is a normal gameDelegate that should also have its ConfigureEnums output.
type secondGameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (s *secondGameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return nil
}

//This delegate already has a manual configureEnums, so shouldn't have one
//automatically generated.
type alreadyHasEnumsGameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (a *alreadyHasEnumsGameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return nil
}

func (a *alreadyHasEnumsGameDelegate) ConfigureEnums() *enum.Set {
	//Because we have htis, we shouldn't export an enums.
	return nil
}

//This delegate shouldn't have ConfigureEnums generated because it has
//AnotherMethodName, not ConfigureMoves.
type fakeGameDelegateWrongMethodName struct {
	boardgame.DefaultGameDelegate
}

func (f *fakeGameDelegateWrongMethodName) AnotherMethodName() *boardgame.MoveTypeConfigBundle {
	return nil
}

//This delegate shouldn't have ConfigureEnums generated because the return
//type doesn't match the ConfigureMoves() signature.
type fakeGameDelegateWrongReturnType struct {
	boardgame.DefaultGameDelegate
}

func (f *fakeGameDelegateWrongReturnType) ConfigureMoves() *boardgame.DefaultGameDelegate {
	return nil
}

//+autoreader
const (
	ColorUnknown = iota
	ColorBlue
	ColorGreen
	ColorRed
)

//+autoreader
const (
	PhaseUnknown = iota
	PhaseMultiWord
	PhaseVeryLongName
)

//+autoreader
const (
	//Because we're overriding to "", this will cause the enum to be a TreeEnum.
	//display:""
	FooOverrideBlank = iota
	//This shouldn't have any weird output
	FooBlue
	//This should work even though enum instruction is in second line.
	//display:"Green"
	FooOverride
	//display:"My name is \"Blue\""
	FooOverrideQuoted
)

//+autoreader
const (
	TransformExampleNormalTransform = iota
	//transform: lower
	TransformExampleLowerCase
	//Transform:UPPER
	TransformExampleUpperCase
	//transform:none
	TransformExampleNormalConfiguredTransform
)

//+autoreader
//transform: upper
const (
	DefaultTransformBlue = iota
	DefaultTransformGreen
	//transform:none
	DefaultTransformRed
)

const (
	DontIncludeBlue = iota
	DontIncludeGreen
)

//+autoreader
const (
	Tree = iota
	TreeBlue
	TreeGreen
	TreeRed
)

//+autoreader
const (
	Blam = iota
	BlamOne
	BlamTwo
	BlamThree
	BlamOne_One
	BlamOne_Two
	BlamTwo_One
)

//+autoreader
const (
	Example = iota
	ExampleOne
	ExampleTwo
	//display:"One > One"
	ExampleOneOne
	ExampleOne_Two
)

//+autoreader
const (
	SkipNode = iota
	//SkipNodeRed is implied but not listed
	SkipNodeRed_Circle
)

//+autoreader all
type myStruct struct {
	boardgame.BaseSubState
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

//+autoreader
type roundRobinStruct struct {
	roundrobinhelpers.BaseGameState
	MyBool bool
}

/*

Long comment

+autoreader
*/
type structWithManyKeys struct {
	boardgame.BaseSubState
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

//+autoreader
type embeddedStruct struct {
	moves.CurrentPlayer
	MyInt int
}

//+autoreader
type doubleEmbeddedStruct struct {
	embeddedStruct
}

//	 +autoreader
type myOtherStruct struct {
	blarg           int
	MyGrowableStack boardgame.Stack
	ThePlayerIndex  boardgame.PlayerIndex
}

type noReaderStruct struct {
	MyInt int
}

// +autoreader reader
type onlyReader struct {
	MyString string
}

//+autoreader
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

// +autoreader    readSetter
type upToReadSetter struct {
	MyInt int
}

//+autoreader
type sizedStackExample struct {
	MySizedStack        boardgame.ImmutableSizedStack
	MyMutableSizedStack boardgame.SizedStack
}

//+autoreader
type mergedStackExample struct {
	MyMergedStack boardgame.MergedStack
}

//+autoreader
type rangeValExample struct {
	MyMutableRangeVal enum.RangeVal
	MyRangeVal        enum.ImmutableRangeVal
}

//+autoreader
type treeValExample struct {
	MyTreeVal          enum.TreeVal
	MyImmutableTreeVal enum.ImmutableTreeVal
}
