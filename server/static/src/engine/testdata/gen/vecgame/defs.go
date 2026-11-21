package vecgame

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/jkomoros/boardgame"
)

// Definition is the engine's GameDefinition (see ../../../definition.ts),
// derived from a live GameManager rather than written by hand: property
// types, stack shapes and struct-tag sanitization policies are read back from
// the manager the Go core actually built.
type Definition struct {
	Name                   string
	Decks                  map[string]DeckDef
	Enums                  map[string]EnumDef
	GroupEnum              string
	Game                   SubStateDef
	Player                 SubStateDef
	DynamicComponentValues map[string]SubStateDef
}

type DeckDef struct {
	Size       int
	Components []json.RawMessage
}

type EnumDef struct {
	// Values maps the decimal enum key to the string a value serializes as.
	Values       map[string]string
	DefaultValue int
}

type SubStateDef struct {
	Props map[string]PropDef
}

type PropDef struct {
	Type string
	// Sanitize maps group name to policy name, exactly as Go parsed the
	// struct tag (StructInflater.PropertySanitizationPolicy).
	Sanitize map[string]string
	Enum     string `json:",omitempty"`
	Deck     string `json:",omitempty"`
	// Size is the slot count of a sized stack, or the space count of a board.
	Size    int `json:",omitempty"`
	MaxSize int `json:",omitempty"`
	// Merged stacks: the sibling properties they view, in order.
	Overlap bool     `json:",omitempty"`
	Stacks  []string `json:",omitempty"`
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
}

var policyNames = map[boardgame.Policy]string{
	boardgame.PolicyVisible:  "visible",
	boardgame.PolicyOrder:    "order",
	boardgame.PolicyLen:      "len",
	boardgame.PolicyNonEmpty: "nonempty",
	boardgame.PolicyHidden:   "hidden",
	boardgame.PolicyInvalid:  "invalid",
}

// PolicyName is the lowercase name the engine uses for a policy.
func PolicyName(p boardgame.Policy) string { return policyNames[p] }

func subStateDef(manager *boardgame.GameManager, sub boardgame.ImmutableSubState, ref boardgame.StatePropertyRef) (SubStateDef, error) {
	result := SubStateDef{Props: map[string]PropDef{}}
	reader := sub.Reader()
	inflater := manager.Internals().StructInflater(ref)

	// Physical stacks by identity, so a merged stack can name its parts.
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
		def := PropDef{Sanitize: map[string]string{}}
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
			} else {
				def.Type = "stack"
				def.MaxSize = stack.MaxSize()
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

// DefinitionFor derives the engine definition for any game's manager.
func DefinitionFor(manager *boardgame.GameManager) (*Definition, error) {
	result := &Definition{
		Name:                   manager.Delegate().Name(),
		Decks:                  map[string]DeckDef{},
		Enums:                  map[string]EnumDef{},
		DynamicComponentValues: map[string]SubStateDef{},
	}

	chest := manager.Chest()
	for _, deckName := range chest.DeckNames() {
		deck := chest.Deck(deckName)
		def := DeckDef{Size: deck.Len(), Components: []json.RawMessage{}}
		for _, c := range deck.Components() {
			blob, err := json.Marshal(c.Values())
			if err != nil {
				return nil, err
			}
			def.Components = append(def.Components, blob)
		}
		result.Decks[deckName] = def
	}

	names := chest.Enums().EnumNames()
	sort.Strings(names)
	for _, name := range names {
		e := chest.Enums().Enum(name)
		def := EnumDef{Values: map[string]string{}, DefaultValue: int(e.DefaultValue())}
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
	result.Game, err = subStateDef(manager, example.ImmutableGameState(), boardgame.StatePropertyRef{Group: boardgame.StateGroupGame})
	if err != nil {
		return nil, err
	}
	result.Player, err = subStateDef(manager, example.ImmutablePlayerStates()[0], boardgame.StatePropertyRef{Group: boardgame.StateGroupPlayer})
	if err != nil {
		return nil, err
	}
	for deckName, values := range example.ImmutableDynamicComponentValues() {
		if len(values) == 0 {
			continue
		}
		result.DynamicComponentValues[deckName], err = subStateDef(manager, values[0], boardgame.StatePropertyRef{Group: boardgame.StateGroupDynamicComponentValues, DeckName: deckName})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
