package examplepkg

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestEnum(t *testing.T) {
	assert.For(t).ThatActual(Enums).IsNotNil()
	assert.For(t).ThatActual(ColorEnum).IsNotNil()

	assert.For(t).ThatActual(ColorEnum.ValueFromString("Unknown")).Equals(ColorUnknown)
	assert.For(t).ThatActual(ColorEnum.ValueFromString("Red")).Equals(ColorRed)
	assert.For(t).ThatActual(ColorEnum.ValueFromString("Green")).Equals(ColorGreen)
	assert.For(t).ThatActual(ColorEnum.ValueFromString("Blue")).Equals(ColorBlue)

}

func TestMain(t *testing.T) {
	var readerObj boardgame.Reader

	obj := &myStruct{}

	readerObj = obj

	reader := readerObj.Reader()

	assert.For(t).ThatActual(reader).IsNotNil()

	obj.MyBool = true

	val, err := reader.BoolProp("MyBool")

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(val).IsTrue()

	var readSetterObj boardgame.ReadSetter

	//This will fail to compile if obj doesn't implement ReadSetter()
	readSetterObj = obj

	readSetter := readSetterObj.ReadSetter()

	assert.For(t).ThatActual(readSetter).IsNotNil()

	err = readSetter.SetBoolProp("MyBool", false)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(obj.MyBool).IsFalse()

}
