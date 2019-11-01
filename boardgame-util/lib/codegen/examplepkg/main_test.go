package examplepkg

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
)

func TestEnum(t *testing.T) {
	assert.For(t).ThatActual(enums).IsNotNil()
	assert.For(t).ThatActual(colorEnum).IsNotNil()

	assert.For(t).ThatActual(colorEnum.ValueFromString("Unknown")).Equals(colorUnknown)
	assert.For(t).ThatActual(colorEnum.ValueFromString("Red")).Equals(colorRed)
	assert.For(t).ThatActual(colorEnum.ValueFromString("Green")).Equals(colorGreen)
	assert.For(t).ThatActual(colorEnum.ValueFromString("Blue")).Equals(colorBlue)

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

	err = readSetter.SetProp("MyInt", 3)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(obj.MyInt).Equals(3)

}
