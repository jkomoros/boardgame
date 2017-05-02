/*

	This is just an example package for testing. Note that some of the uses of
	autoreader use somewhat odd spacing or capitalization; this is primarily
	just to test how resilient the package is to unexpected input.

*/
package examplepkg

import (
	"github.com/jkomoros/boardgame"
)

//go:generate autoreader

//+autoreader both
type myStruct struct {
	MyInt              int
	MyBool             bool
	MySizedStack       *boardgame.SizedStack
	TheTimer           *boardgame.Timer
	MyIntSlice         []int
	MyBoolSlice        []bool
	MyStringSlice      []string
	MyPlayerIndexSlice []boardgame.PlayerIndex
}

//	 +autoreader
type myOtherStruct struct {
	blarg           int
	MyGrowableStack *boardgame.GrowableStack
	ThePlayerIndex  boardgame.PlayerIndex
}

type noReaderStruct struct {
	MyInt int
}

// +autoreader reader
type onlyReader struct {
	MyString string
}

// +autoreader    readSetter
type onlyReadSetter struct {
	MyInt int
}
