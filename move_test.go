package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestStructTag(t *testing.T) {

	type anonTestStruct struct {
		C int `enum:"C"`
	}

	type anonPointerTestStruct struct {
		D int `enum:"D"`
	}

	type testStruct struct {
		anonTestStruct
		*anonPointerTestStruct
		A int
		B enum.Var `enum:"B"`
	}

	theStruct := &testStruct{
		anonPointerTestStruct: &anonPointerTestStruct{},
	}

	assert.For(t).ThatActual(enumStructTagForField(theStruct, "A")).Equals("")
	assert.For(t).ThatActual(enumStructTagForField(theStruct, "B")).Equals("B")
	assert.For(t).ThatActual(enumStructTagForField(theStruct, "C")).Equals("C")
	assert.For(t).ThatActual(enumStructTagForField(theStruct, "D")).Equals("D")
	assert.For(t).ThatActual(enumStructTagForField(theStruct, "Illegal")).Equals("")

}
