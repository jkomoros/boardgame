package gamedef

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// typeNames is the engine's PropType for each Go property type, as
// definition.ts spells it. Stacks and boards are named by subStateDef from
// their shape instead.
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

// TypeName is the engine's PropType name for a scalar or slice property
// type ("int", "playerIndex", "enumSlice", ...), or "" for a stack.
func TypeName(t boardgame.PropertyType) string { return typeNames[t] }

// DefinitionFor derives the engine definition (definition.ts's
// GameDefinition) from a live manager: property types, stack shapes and
// struct-tag sanitization policies are read back from what the Go core
// actually built. It includes the additions GenericValues, ComponentProps,
// Enums[].Parents, Enums[].Range, PropDef.Constraints, and the enum of a
// board keyed by an enum; WithoutAdditions removes them.
func DefinitionFor(manager *boardgame.GameManager) (*Definition, error) {
	if manager == nil {
		return nil, errors.New("gamedef: nil manager")
	}
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
		var example boardgame.ComponentValues
		for _, c := range deck.Components() {
			blob, err := json.Marshal(c.Values())
			if err != nil {
				return nil, err
			}
			def.Components = append(def.Components, blob)
			if example == nil && c.Values() != nil {
				example = c.Values()
			}
		}
		if generic := deck.GenericComponent().Values(); generic != nil {
			blob, err := json.Marshal(generic)
			if err != nil {
				return nil, err
			}
			def.GenericValues = blob
		}
		if example != nil {
			props, err := componentProps(example.Reader())
			if err != nil {
				return nil, errors.New("deck " + deckName + ": " + err.Error())
			}
			def.ComponentProps = props
		}
		result.Decks[deckName] = def
	}

	for _, name := range chest.Enums().EnumNames() {
		result.Enums[name] = enumDef(chest.Enums().Enum(name))
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

// WithoutAdditions is d without the fields DefinitionFor emits beyond what
// server/static/src/engine/testdata/gen/vecgame emitted before them
// (GenericValues, ComponentProps, Parents, Range, Constraints, and a board's
// Enum). It serializes byte for byte as the checked-in
// testdata/defs/*.def.json. d is not modified.
func (d *Definition) WithoutAdditions() *Definition {
	result := &Definition{
		Name:                   d.Name,
		Decks:                  map[string]DeckDef{},
		Enums:                  map[string]EnumDef{},
		GroupEnum:              d.GroupEnum,
		Game:                   d.Game.withoutAdditions(),
		Player:                 d.Player.withoutAdditions(),
		DynamicComponentValues: map[string]SubStateDef{},
	}
	for name, deck := range d.Decks {
		result.Decks[name] = DeckDef{Size: deck.Size, Components: deck.Components}
	}
	for name, e := range d.Enums {
		result.Enums[name] = EnumDef{Values: e.Values, DefaultValue: e.DefaultValue}
	}
	for name, sub := range d.DynamicComponentValues {
		result.DynamicComponentValues[name] = sub.withoutAdditions()
	}
	return result
}

func (s SubStateDef) withoutAdditions() SubStateDef {
	result := SubStateDef{Props: map[string]PropDef{}}
	for name, prop := range s.Props {
		prop.Constraints = nil
		if prop.Type == "board" {
			prop.Enum = ""
		}
		result.Props[name] = prop
	}
	return result
}

// enumDef describes one enum, with a tree enum's parents and a range enum's
// dimensions.
func enumDef(e enum.Enum) EnumDef {
	def := EnumDef{Values: map[string]string{}, DefaultValue: int(e.DefaultValue())}
	for _, key := range e.Values() {
		def.Values[strconv.Itoa(int(key))] = e.String(key)
	}
	if tree := e.TreeEnum(); tree != nil {
		def.Parents = map[string]int{}
		for _, key := range e.Values() {
			def.Parents[strconv.Itoa(int(key))] = int(tree.Parent(key))
		}
	}
	if ranged := e.RangeEnum(); ranged != nil {
		def.Range = append([]int{}, ranged.RangeDimensions()...)
	}
	return def
}

// componentProps is each component value property's type and enum.
func componentProps(reader boardgame.PropertyReader) (map[string]ComponentPropDef, error) {
	result := map[string]ComponentPropDef{}
	for name, propType := range reader.Props() {
		typeName, ok := typeNames[propType]
		if !ok {
			return nil, errors.New(name + " has an unsupported component value type")
		}
		def := ComponentPropDef{Type: typeName}
		switch propType {
		case boardgame.TypeEnum:
			val, err := reader.ImmutableEnumProp(name)
			if err != nil {
				return nil, err
			}
			if val != nil && val.Enum() != nil {
				def.Enum = val.Enum().Name()
			}
		case boardgame.TypeEnumSlice:
			val, err := reader.ImmutableEnumSliceProp(name)
			if err != nil {
				return nil, err
			}
			if val != nil && val.Enum() != nil {
				def.Enum = val.Enum().Name()
			}
		}
		result[name] = def
	}
	return result, nil
}

// componentEnums are the enums component values (and dynamic values) use
// that the chest's enum set does not hold.
func componentEnums(manager *boardgame.GameManager) map[string]EnumDef {
	result := map[string]EnumDef{}
	chest := manager.Chest()
	add := func(reader boardgame.PropertyReader) {
		for name, propType := range reader.Props() {
			var e enum.Enum
			switch propType {
			case boardgame.TypeEnum:
				if val, err := reader.ImmutableEnumProp(name); err == nil && val != nil {
					e = val.Enum()
				}
			case boardgame.TypeEnumSlice:
				if val, err := reader.ImmutableEnumSliceProp(name); err == nil && val != nil {
					e = val.Enum()
				}
			}
			if e == nil || chest.Enums().Enum(e.Name()) == e {
				continue
			}
			result[e.Name()] = enumDef(e)
		}
	}
	for _, deckName := range chest.DeckNames() {
		for _, c := range chest.Deck(deckName).Components() {
			if c.Values() != nil {
				add(c.Values().Reader())
			}
		}
	}
	if example := manager.ExampleState(); example != nil {
		for _, values := range example.ImmutableDynamicComponentValues() {
			if len(values) > 0 {
				add(values[0].Reader())
			}
		}
	}
	return result
}

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
			def.Constraints = stackConstraints(sub, name)
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
			if e := board.Enum(); e != nil {
				def.Enum = e.Name()
			}
			def.Constraints = stackConstraints(sub, name)
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

// stackConstraints are the constraint expressions ("same(Color)") in the
// stack or sizedstack struct tag of the field name on sub's concrete struct,
// in tag order, split as the core's struct inflater splits them.
func stackConstraints(sub interface{}, name string) []string {
	t := reflect.TypeOf(sub)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	field, ok := t.FieldByName(name)
	if !ok {
		return nil
	}
	tag := field.Tag.Get("stack")
	if tag == "" {
		tag = field.Tag.Get("sizedstack")
	}
	var result []string
	for i, piece := range splitTagFields(tag) {
		piece = strings.TrimSpace(piece)
		if i == 0 || piece == "" {
			continue
		}
		if isConstraintExpr(piece) {
			result = append(result, piece)
		}
	}
	return result
}

// splitTagFields is the core's struct_inflater.go splitTagFields: split on
// commas outside parentheses.
func splitTagFields(tag string) []string {
	if tag == "" {
		return nil
	}
	var fields []string
	depth, start := 0, 0
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

// isConstraintExpr is the shape the core's parseConstraintExpr accepts:
// name(args), with a non-empty name.
func isConstraintExpr(s string) bool {
	open := strings.IndexByte(s, '(')
	return open > 0 && strings.HasSuffix(s, ")")
}

// sortedKeys returns m's keys in order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
