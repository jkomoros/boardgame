package main

// The engine GameDefinition (../../../definition.ts) for a live manager,
// with each stack's constraint strings. This is the shape
// ../vecgame/defs.go writes, plus PropDef.Constraints and
// DeckDef.ComponentProps. It is kept here rather than shared so that the
// constraint vectors do not depend on that file, which is being moved to
// boardgame-util/lib/gamedef.

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
)

type definition struct {
	Name                   string
	Decks                  map[string]deckDef
	Enums                  map[string]enumDef
	GroupEnum              string
	Game                   subStateDef
	Player                 subStateDef
	DynamicComponentValues map[string]subStateDef
}

type deckDef struct {
	Size           int
	Components     []json.RawMessage
	ComponentProps map[string]componentPropDef `json:",omitempty"`
}

type componentPropDef struct {
	Type string
	Enum string `json:",omitempty"`
}

type enumDef struct {
	Values       map[string]string
	DefaultValue int
}

type subStateDef struct {
	Props map[string]propDef
}

type propDef struct {
	Type        string
	Sanitize    map[string]string
	Enum        string   `json:",omitempty"`
	Deck        string   `json:",omitempty"`
	Size        int      `json:",omitempty"`
	MaxSize     int      `json:",omitempty"`
	Overlap     bool     `json:",omitempty"`
	Stacks      []string `json:",omitempty"`
	Constraints []string `json:",omitempty"`
}

var typeNames = map[boardgame.PropertyType]string{
	boardgame.TypeInt:              "int",
	boardgame.TypeBool:             "bool",
	boardgame.TypeString:           "string",
	boardgame.TypePlayerIndex:      "playerIndex",
	boardgame.TypeEnum:             "enum",
	boardgame.TypeIntSlice:         "intSlice",
	boardgame.TypeBoolSlice:        "boolSlice",
	boardgame.TypeStringSlice:      "stringSlice",
	boardgame.TypePlayerIndexSlice: "playerIndexSlice",
	boardgame.TypeTimer:            "timer",
	boardgame.TypeEnumSlice:        "enumSlice",
	boardgame.TypeBoard:            "board",
	boardgame.TypeStack:            "stack",
}

var policyNames = map[boardgame.Policy]string{
	boardgame.PolicyVisible:  "visible",
	boardgame.PolicyOrder:    "order",
	boardgame.PolicyLen:      "len",
	boardgame.PolicyNonEmpty: "nonempty",
	boardgame.PolicyHidden:   "hidden",
	boardgame.PolicyInvalid:  "invalid",
}

// splitTagFields and isConstraintExpr follow struct_inflater.go's
// splitTagFields and parseConstraintExpr, which are unexported: a top-level
// comma split that ignores commas inside parentheses, and "name(args)".
func splitTagFields(tag string) []string {
	var fields []string
	depth := 0
	start := 0
	for i, c := range tag {
		switch c {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				fields = append(fields, tag[start:i])
				start = i + 1
			}
		}
	}
	return append(fields, tag[start:])
}

func isConstraintExpr(s string) bool {
	paren := strings.Index(s, "(")
	return paren > 0 && strings.HasSuffix(s, ")")
}

// constraintsFromTag reads the constraint pieces of the stack or sizedstack
// struct tag of field name on the concrete sub-state, trimmed, in tag order.
func constraintsFromTag(sub interface{}, name string) []string {
	t := reflect.TypeOf(sub)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, ok := t.FieldByName(name)
	if !ok {
		return nil
	}
	tag := field.Tag.Get("stack")
	if tag == "" {
		tag = field.Tag.Get("sizedstack")
	}
	if tag == "" {
		return nil
	}
	var out []string
	for i, piece := range splitTagFields(tag) {
		piece = strings.TrimSpace(piece)
		if i == 0 || piece == "" || !isConstraintExpr(piece) {
			continue
		}
		out = append(out, piece)
	}
	return out
}

func subStateDefFor(manager *boardgame.GameManager, sub boardgame.ImmutableSubState, ref boardgame.StatePropertyRef) (subStateDef, error) {
	result := subStateDef{Props: map[string]propDef{}}
	reader := sub.Reader()
	inflater := manager.Internals().StructInflater(ref)

	physical := map[boardgame.ImmutableStack]string{}
	for name, propType := range reader.Props() {
		if propType != boardgame.TypeStack {
			continue
		}
		stack, err := reader.ImmutableStackProp(name)
		if err != nil {
			return result, err
		}
		physical[stack] = name
	}

	for name, propType := range reader.Props() {
		def := propDef{Sanitize: map[string]string{}}
		if inflater != nil {
			for group, policy := range inflater.PropertySanitizationPolicy(name) {
				def.Sanitize[group] = policyNames[policy]
			}
		}
		switch propType {
		case boardgame.TypeStack:
			stack, err := reader.ImmutableStackProp(name)
			if err != nil {
				return result, err
			}
			def.Deck = stack.Deck().Name()
			if merged := stack.MergedStack(); merged != nil {
				def.Type = "mergedStack"
				def.Overlap = merged.Overlapped()
				for _, part := range merged.ImmutableStacks() {
					partName, ok := physical[part]
					if !ok {
						return result, errors.New(name + " merges a stack that is not a sibling property")
					}
					def.Stacks = append(def.Stacks, partName)
				}
			} else if stack.ImmutableSizedStack() != nil {
				def.Type = "sizedStack"
				def.Size = stack.Len()
				def.Constraints = constraintsFromTag(sub, name)
			} else {
				def.Type = "stack"
				def.MaxSize = stack.MaxSize()
				def.Constraints = constraintsFromTag(sub, name)
			}
		case boardgame.TypeBoard:
			board, err := reader.ImmutableBoardProp(name)
			if err != nil {
				return result, err
			}
			def.Type = "board"
			def.Size = board.Len()
			if board.Len() > 0 {
				def.Deck = board.ImmutableSpaceAt(0).Deck().Name()
				def.MaxSize = board.ImmutableSpaceAt(0).MaxSize()
			}
			def.Constraints = constraintsFromTag(sub, name)
		case boardgame.TypeEnum:
			val, err := reader.ImmutableEnumProp(name)
			if err != nil {
				return result, err
			}
			def.Type = "enum"
			def.Enum = val.Enum().Name()
		case boardgame.TypeEnumSlice:
			val, err := reader.ImmutableEnumSliceProp(name)
			if err != nil {
				return result, err
			}
			def.Type = "enumSlice"
			def.Enum = val.Enum().Name()
		default:
			typeName, ok := typeNames[propType]
			if !ok {
				return result, errors.New(name + " has an unsupported property type")
			}
			def.Type = typeName
		}
		result.Props[name] = def
	}
	return result, nil
}

func definitionFor(manager *boardgame.GameManager) (*definition, error) {
	result := &definition{
		Name:                   manager.Delegate().Name(),
		Decks:                  map[string]deckDef{},
		Enums:                  map[string]enumDef{},
		DynamicComponentValues: map[string]subStateDef{},
	}

	chest := manager.Chest()
	for _, deckName := range chest.DeckNames() {
		deck := chest.Deck(deckName)
		def := deckDef{Size: deck.Len(), Components: []json.RawMessage{}}
		for i, c := range deck.Components() {
			blob, err := json.Marshal(c.Values())
			if err != nil {
				return nil, err
			}
			def.Components = append(def.Components, blob)
			if i != 0 || c.Values() == nil || c.Values().Reader() == nil {
				continue
			}
			reader := c.Values().Reader()
			def.ComponentProps = map[string]componentPropDef{}
			for name, propType := range reader.Props() {
				prop := componentPropDef{Type: typeNames[propType]}
				switch propType {
				case boardgame.TypeEnum:
					if val, err := reader.ImmutableEnumProp(name); err == nil && val != nil {
						prop.Enum = val.Enum().Name()
					}
				case boardgame.TypeEnumSlice:
					if val, err := reader.ImmutableEnumSliceProp(name); err == nil && val != nil {
						prop.Enum = val.Enum().Name()
					}
				}
				def.ComponentProps[name] = prop
			}
		}
		result.Decks[deckName] = def
	}

	names := chest.Enums().EnumNames()
	sort.Strings(names)
	for _, name := range names {
		e := chest.Enums().Enum(name)
		def := enumDef{Values: map[string]string{}, DefaultValue: int(e.DefaultValue())}
		for _, key := range e.Values() {
			def.Values[strconv.Itoa(int(key))] = e.String(key)
		}
		result.Enums[name] = def
	}
	if group := manager.Delegate().GroupEnum(); group != nil {
		result.GroupEnum = group.Name()
	}

	example := manager.ExampleState()
	if example == nil {
		return nil, errors.New("no example state")
	}
	var err error
	result.Game, err = subStateDefFor(manager, example.ImmutableGameState(), boardgame.StatePropertyRef{Group: boardgame.StateGroupGame})
	if err != nil {
		return nil, err
	}
	result.Player, err = subStateDefFor(manager, example.ImmutablePlayerStates()[0], boardgame.StatePropertyRef{Group: boardgame.StateGroupPlayer})
	if err != nil {
		return nil, err
	}
	for deckName, values := range example.ImmutableDynamicComponentValues() {
		if len(values) == 0 {
			continue
		}
		result.DynamicComponentValues[deckName], err = subStateDefFor(manager, values[0], boardgame.StatePropertyRef{Group: boardgame.StateGroupDynamicComponentValues, DeckName: deckName})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
