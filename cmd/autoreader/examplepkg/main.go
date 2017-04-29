/*
	This is just an example package for testing
*/
package examplepkg

import (
	"github.com/jkomoros/boardgame"
)

//go:generate autoreader

//+autoreader
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
