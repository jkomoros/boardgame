/************************************
 *
 * This file contains auto-generated methods to help configure enums.
 * It was generated by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package tictactoe

import (
	"github.com/jkomoros/boardgame/enum"
)

var enums = enum.NewSet()

//ConfigureEnums simply returns enums, the auto-generated Enums variable. This
//is output because gameDelegate appears to be a struct that implements
//boardgame.GameDelegate, and does not already have a ConfigureEnums
//explicitly defined.
func (g *gameDelegate) ConfigureEnums() *enum.Set {
	return enums
}

//phaseEnum is the enum.Enum for phase
var phaseEnum = enums.MustAdd("phase", map[int]string{
	phaseAfterFirstMove:  "After First Move",
	phaseBeforeFirstMove: "Before First Move",
})
