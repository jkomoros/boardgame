package examplepkg

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestMain(t *testing.T) {
	var mutableSubState boardgame.MutableSubState

	obj := &myStruct{}

	//This will fail to compile if obj doesn't implement Reader()
	mutableSubState = obj

	reader := mutableSubState.Reader()

	assert.For(t).ThatActual(reader).IsNotNil()

	obj.MyBool = true

	val, err := reader.BoolProp("MyBool")

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(val).IsTrue()

	readSetter := mutableSubState.ReadSetter()

	assert.For(t).ThatActual(readSetter).IsNotNil()

	err = readSetter.SetBoolProp("MyBool", false)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(obj.MyBool).IsFalse()

}
