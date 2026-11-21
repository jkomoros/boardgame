package vecgame

import (
	"errors"
	"fmt"

	"github.com/jkomoros/boardgame"
)

// StackRef names one physical stack the way the engine's StackRef does.
type StackRef struct {
	Group  string // "game", "player" or "dynamic"
	Player int    `json:",omitempty"`
	Deck   string `json:",omitempty"`
	Index  int    `json:",omitempty"`
	Prop   string
	Space  *int `json:",omitempty"`
}

// Path is the owner path stack_ownership.go gives the stack.
func (r StackRef) Path() string {
	var base string
	switch r.Group {
	case "game":
		base = "Game"
	case "player":
		base = fmt.Sprintf("Players[%d]", r.Player)
	default:
		base = fmt.Sprintf("DynamicComponentValues[%s][%d]", r.Deck, r.Index)
	}
	path := base + "." + r.Prop
	if r.Space != nil {
		path += fmt.Sprintf("[%d]", *r.Space)
	}
	return path
}

// AllStackRefs lists every physical stack of a state with numPlayers players.
// Only the first three cards' dynamic stacks are listed, to keep vectors small.
func AllStackRefs(numPlayers int) []StackRef {
	var out []StackRef
	for _, prop := range []string{"Draw", "Limited", "Row", "Under", "Tokens", "TagLen", "TagOrder", "TagNonEmpty"} {
		out = append(out, StackRef{Group: "game", Prop: prop})
	}
	for i := 0; i < 3; i++ {
		space := i
		out = append(out, StackRef{Group: "game", Prop: "Spaces", Space: &space})
	}
	for p := 0; p < numPlayers; p++ {
		out = append(out, StackRef{Group: "player", Player: p, Prop: "Hand"})
		out = append(out, StackRef{Group: "player", Player: p, Prop: "Slots"})
	}
	for i := 0; i < 3; i++ {
		out = append(out, StackRef{Group: "dynamic", Deck: "cards", Index: i, Prop: "Held"})
	}
	return out
}

// Resolve finds the live stack a ref names in a mutable state.
func Resolve(st boardgame.State, ref StackRef) (boardgame.Stack, error) {
	switch ref.Group {
	case "game":
		g := Game(st)
		switch ref.Prop {
		case "Draw":
			return g.Draw, nil
		case "Limited":
			return g.Limited, nil
		case "Row":
			return g.Row, nil
		case "Under":
			return g.Under, nil
		case "Tokens":
			return g.Tokens, nil
		case "TagLen":
			return g.TagLen, nil
		case "TagOrder":
			return g.TagOrder, nil
		case "TagNonEmpty":
			return g.TagNonEmpty, nil
		case "Spaces":
			if ref.Space == nil {
				return nil, errors.New("no space")
			}
			return g.Spaces.SpaceAt(*ref.Space), nil
		}
	case "player":
		p := Players(st)[ref.Player]
		switch ref.Prop {
		case "Hand":
			return p.Hand, nil
		case "Slots":
			return p.Slots, nil
		}
	case "dynamic":
		if ref.Prop == "Held" {
			return CardDynamic(st, ref.Index).Held, nil
		}
	}
	return nil, errors.New("unknown stack ref " + ref.Path())
}
