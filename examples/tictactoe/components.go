package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const (
	X     = "X"
	O     = "O"
	Empty = ""
)

//+autoreader reader
type playerToken struct {
	Value string
}

//Designed to be used with stack.ComponentValues()
func playerTokenValues(in []boardgame.SubState) []*playerToken {
	result := make([]*playerToken, len(in))
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c == nil {
			result[i] = nil
			continue
		}
		result[i] = c.(*playerToken)
	}
	return result
}

const (
	ColorUnknown = iota
	ColorRed
	ColorGreen
)

type EnumManager struct {
	myInt int
}

type Enum struct {
	Name    string
	Manager *EnumManager
}

func NewEnumManager() *EnumManager {
	return &EnumManager{}
}

func (e *EnumManager) NewEnum(name string, values map[int]string) *Enum {
	return &Enum{name, e}
}

var enumManager = NewEnumManager()

var ColorEnum = enumManager.NewEnum("Color", map[int]string{
	ColorUnknown: "Unknown",
	ColorRed:     "Red",
	ColorGreen:   "Green",
})
