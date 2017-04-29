/*
	This is just an example package for testing
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
