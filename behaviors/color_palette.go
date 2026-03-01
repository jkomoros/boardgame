package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Named color constants for game color enums, ordered by frequency of use in
// board games. Use these as enum key values when defining your "color" enum.
// Each constant maps to a recognizable CSS color.
const (
	ColorRed    enum.EnumKey = iota // #D32F2F
	ColorBlue                       // #1976D2
	ColorGreen                      // #388E3C
	ColorYellow                     // #FBC02D
	ColorBlack                      // #424242
	ColorWhite                      // #FAFAFA
	ColorOrange                     // #E64A19
	ColorPurple                     // #7B1FA2
	ColorPink                       // #C2185B
	ColorBrown                      // #795548
	ColorCyan                       // #0097A7
	ColorGray                       // #757575
)

// CSSColorForKey maps each named color constant to its CSS color string.
var CSSColorForKey = map[enum.EnumKey]string{
	ColorRed:    "#D32F2F",
	ColorBlue:   "#1976D2",
	ColorGreen:  "#388E3C",
	ColorYellow: "#FBC02D",
	ColorBlack:  "#424242",
	ColorWhite:  "#FAFAFA",
	ColorOrange: "#E64A19",
	ColorPurple: "#7B1FA2",
	ColorPink:   "#C2185B",
	ColorBrown:  "#795548",
	ColorCyan:   "#0097A7",
	ColorGray:   "#757575",
}

// defaultPlayerPalette is the order used for player-index-based fallback.
// Picks maximally distinct colors first.
var defaultPlayerPalette = []string{
	CSSColorForKey[ColorRed],
	CSSColorForKey[ColorBlue],
	CSSColorForKey[ColorGreen],
	CSSColorForKey[ColorYellow],
	CSSColorForKey[ColorOrange],
	CSSColorForKey[ColorPurple],
	CSSColorForKey[ColorCyan],
	CSSColorForKey[ColorPink],
	CSSColorForKey[ColorBlack],
	CSSColorForKey[ColorBrown],
	CSSColorForKey[ColorGray],
	CSSColorForKey[ColorWhite],
}

// DefaultPlayerColor returns the CSS color for a player index, cycling
// through the default palette.
func DefaultPlayerColor(index int) string {
	if index < 0 {
		index = 0
	}
	return defaultPlayerPalette[index%len(defaultPlayerPalette)]
}

// CSSColorForPlayer returns a CSS color string for the given player. If the
// playerState has a Color enum property (from PlayerColor behavior), looks up
// the enum key value in the named constants. Falls back to player index in the
// default palette.
func CSSColorForPlayer(playerState boardgame.ImmutableSubState) string {
	colorVal, err := playerState.Reader().ImmutableEnumProp(colorPropertyName)
	if err == nil && colorVal != nil {
		if css, ok := CSSColorForKey[colorVal.Value()]; ok {
			return css
		}
	}
	playerIndex := int(playerState.StatePropertyRef().PlayerIndex)
	return DefaultPlayerColor(playerIndex)
}
