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
	MyInt  int
	MyBool bool
}

//	 +autoreader
type myOtherStruct struct {
	blarg int
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
