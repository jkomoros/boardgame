/************************************
 *
 * This file contains auto-generated methods to help configure enums.
 * It was generated by autoreader.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package examplepkg

import (
	"github.com/jkomoros/boardgame/enum"
)

var Enums = enum.NewSet()

//ConfigureEnums simply returns Enums, the auto-generated Enums variable. This
//is output because gameDelegate appears to be a struct that implements
//boardgame.GameDelegate, and does not already have a ConfigureEnums
//explicitly defined.
func (g *gameDelegate) ConfigureEnums() *enum.Set {
	return Enums
}

//ConfigureEnums simply returns Enums, the auto-generated Enums variable. This
//is output because secondGameDelegate appears to be a struct that implements
//boardgame.GameDelegate, and does not already have a ConfigureEnums
//explicitly defined.
func (s *secondGameDelegate) ConfigureEnums() *enum.Set {
	return Enums
}

var ColorEnum = Enums.MustAdd("Color", map[int]string{
	ColorBlue:    "Blue",
	ColorGreen:   "Green",
	ColorRed:     "Red",
	ColorUnknown: "Unknown",
})

var PhaseEnum = Enums.MustAdd("Phase", map[int]string{
	PhaseMultiWord:    "Multi Word",
	PhaseUnknown:      "Unknown",
	PhaseVeryLongName: "Very Long Name",
})

var FooEnum = Enums.MustAddTree("Foo", map[int]string{
	FooBlue:           "Blue",
	FooOverride:       "Green",
	FooOverrideBlank:  "",
	FooOverrideQuoted: "My name is \"Blue\"",
}, map[int]int{
	FooBlue:           FooOverrideBlank,
	FooOverride:       FooOverrideBlank,
	FooOverrideBlank:  FooOverrideBlank,
	FooOverrideQuoted: FooOverrideBlank,
})

var TransformExampleEnum = Enums.MustAdd("TransformExample", map[int]string{
	TransformExampleLowerCase:                 "lower case",
	TransformExampleNormalConfiguredTransform: "Normal Configured Transform",
	TransformExampleNormalTransform:           "Normal Transform",
	TransformExampleUpperCase:                 "UPPER CASE",
})

var DefaultTransformEnum = Enums.MustAdd("DefaultTransform", map[int]string{
	DefaultTransformBlue:  "BLUE",
	DefaultTransformGreen: "GREEN",
	DefaultTransformRed:   "Red",
})

var TreeEnum = Enums.MustAddTree("Tree", map[int]string{
	Tree:      "",
	TreeBlue:  "Blue",
	TreeGreen: "Green",
	TreeRed:   "Red",
}, map[int]int{
	Tree:      Tree,
	TreeBlue:  Tree,
	TreeGreen: Tree,
	TreeRed:   Tree,
})
