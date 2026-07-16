package boardgame

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame/enum"
)

// ComputedPropertyScope identifies where a configured computed value appears
// in the client snapshot.
type ComputedPropertyScope string

const (
	ComputedPropertyScopeGlobal ComputedPropertyScope = "global"
	ComputedPropertyScopePlayer ComputedPropertyScope = "player"
)

// ComputedPropertyDescriptor is the immutable, generator-facing description
// of one game-configured computed value.
type ComputedPropertyDescriptor struct {
	Name     string
	Scope    ComputedPropertyScope
	Type     PropertyType
	EnumName string
}

// ComputedProperty couples a computed value's declaration to the callback that
// evaluates it. Its fields are intentionally private; use the typed
// GlobalComputed*/PlayerComputed* constructors.
type ComputedProperty struct {
	descriptor     ComputedPropertyDescriptor
	global         func(ImmutableState) interface{}
	player         func(ImmutableSubState) interface{}
	configuredEnum enum.Enum
	configErr      error
}

func globalComputed(name string, propType PropertyType, enumName string, fn func(ImmutableState) interface{}) ComputedProperty {
	return ComputedProperty{
		descriptor: ComputedPropertyDescriptor{Name: name, Scope: ComputedPropertyScopeGlobal, Type: propType, EnumName: enumName},
		global:     fn,
	}
}

func playerComputed(name string, propType PropertyType, enumName string, fn func(ImmutableSubState) interface{}) ComputedProperty {
	return ComputedProperty{
		descriptor: ComputedPropertyDescriptor{Name: name, Scope: ComputedPropertyScopePlayer, Type: propType, EnumName: enumName},
		player:     fn,
	}
}

// GlobalComputedBool declares an always-present global boolean.
func GlobalComputedBool(name string, fn func(ImmutableState) bool) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeBool, "", func(state ImmutableState) interface{} { return fn(state) })
}

// GlobalComputedInt declares an always-present global integer.
func GlobalComputedInt(name string, fn func(ImmutableState) int) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeInt, "", func(state ImmutableState) interface{} { return fn(state) })
}

// GlobalComputedString declares an always-present global string.
func GlobalComputedString(name string, fn func(ImmutableState) string) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeString, "", func(state ImmutableState) interface{} { return fn(state) })
}

// GlobalComputedPlayerIndex declares an always-present global player index.
func GlobalComputedPlayerIndex(name string, fn func(ImmutableState) PlayerIndex) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypePlayerIndex, "", func(state ImmutableState) interface{} { return fn(state) })
}

// GlobalComputedEnum declares an always-present global value from e.
func GlobalComputedEnum(name string, e enum.Enum, fn func(ImmutableState) enum.ImmutableVal) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	if e == nil {
		return ComputedProperty{configErr: errors.New("enum is nil")}
	}
	result := globalComputed(name, TypeEnum, e.Name(), func(state ImmutableState) interface{} {
		return checkedComputedEnum(name, e, fn(state))
	})
	result.configuredEnum = e
	return result
}

// GlobalComputedBoolSlice declares an always-present global boolean slice.
func GlobalComputedBoolSlice(name string, fn func(ImmutableState) []bool) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeBoolSlice, "", func(state ImmutableState) interface{} { return append([]bool{}, fn(state)...) })
}

// GlobalComputedIntSlice declares an always-present global integer slice.
func GlobalComputedIntSlice(name string, fn func(ImmutableState) []int) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeIntSlice, "", func(state ImmutableState) interface{} { return append([]int{}, fn(state)...) })
}

// GlobalComputedStringSlice declares an always-present global string slice.
func GlobalComputedStringSlice(name string, fn func(ImmutableState) []string) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypeStringSlice, "", func(state ImmutableState) interface{} { return append([]string{}, fn(state)...) })
}

// GlobalComputedPlayerIndexSlice declares an always-present global player-index slice.
func GlobalComputedPlayerIndexSlice(name string, fn func(ImmutableState) []PlayerIndex) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return globalComputed(name, TypePlayerIndexSlice, "", func(state ImmutableState) interface{} { return append([]PlayerIndex{}, fn(state)...) })
}

// GlobalComputedEnumSlice declares an always-present global slice from e.
func GlobalComputedEnumSlice(name string, e enum.Enum, fn func(ImmutableState) enum.ImmutableEnumSlice) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	if e == nil {
		return ComputedProperty{configErr: errors.New("enum is nil")}
	}
	result := globalComputed(name, TypeEnumSlice, e.Name(), func(state ImmutableState) interface{} {
		return checkedComputedEnumSlice(name, e, fn(state))
	})
	result.configuredEnum = e
	return result
}

// PlayerComputedBool declares an always-present per-player boolean.
func PlayerComputedBool(name string, fn func(ImmutableSubState) bool) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeBool, "", func(player ImmutableSubState) interface{} { return fn(player) })
}

// PlayerComputedInt declares an always-present per-player integer.
func PlayerComputedInt(name string, fn func(ImmutableSubState) int) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeInt, "", func(player ImmutableSubState) interface{} { return fn(player) })
}

// PlayerComputedString declares an always-present per-player string.
func PlayerComputedString(name string, fn func(ImmutableSubState) string) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeString, "", func(player ImmutableSubState) interface{} { return fn(player) })
}

// PlayerComputedPlayerIndex declares an always-present per-player player index.
func PlayerComputedPlayerIndex(name string, fn func(ImmutableSubState) PlayerIndex) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypePlayerIndex, "", func(player ImmutableSubState) interface{} { return fn(player) })
}

// PlayerComputedEnum declares an always-present per-player value from e.
func PlayerComputedEnum(name string, e enum.Enum, fn func(ImmutableSubState) enum.ImmutableVal) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	if e == nil {
		return ComputedProperty{configErr: errors.New("enum is nil")}
	}
	result := playerComputed(name, TypeEnum, e.Name(), func(player ImmutableSubState) interface{} {
		return checkedComputedEnum(name, e, fn(player))
	})
	result.configuredEnum = e
	return result
}

// PlayerComputedBoolSlice declares an always-present per-player boolean slice.
func PlayerComputedBoolSlice(name string, fn func(ImmutableSubState) []bool) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeBoolSlice, "", func(player ImmutableSubState) interface{} { return append([]bool{}, fn(player)...) })
}

// PlayerComputedIntSlice declares an always-present per-player integer slice.
func PlayerComputedIntSlice(name string, fn func(ImmutableSubState) []int) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeIntSlice, "", func(player ImmutableSubState) interface{} { return append([]int{}, fn(player)...) })
}

// PlayerComputedStringSlice declares an always-present per-player string slice.
func PlayerComputedStringSlice(name string, fn func(ImmutableSubState) []string) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypeStringSlice, "", func(player ImmutableSubState) interface{} { return append([]string{}, fn(player)...) })
}

// PlayerComputedPlayerIndexSlice declares an always-present per-player player-index slice.
func PlayerComputedPlayerIndexSlice(name string, fn func(ImmutableSubState) []PlayerIndex) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	return playerComputed(name, TypePlayerIndexSlice, "", func(player ImmutableSubState) interface{} { return append([]PlayerIndex{}, fn(player)...) })
}

// PlayerComputedEnumSlice declares an always-present per-player slice from e.
func PlayerComputedEnumSlice(name string, e enum.Enum, fn func(ImmutableSubState) enum.ImmutableEnumSlice) ComputedProperty {
	if fn == nil {
		return ComputedProperty{configErr: errors.New("callback is nil")}
	}
	if e == nil {
		return ComputedProperty{configErr: errors.New("enum is nil")}
	}
	result := playerComputed(name, TypeEnumSlice, e.Name(), func(player ImmutableSubState) interface{} {
		return checkedComputedEnumSlice(name, e, fn(player))
	})
	result.configuredEnum = e
	return result
}

func checkedComputedEnum(name string, expected enum.Enum, value enum.ImmutableVal) enum.ImmutableVal {
	if value == nil || value.Enum() != expected {
		panic(fmt.Sprintf("computed property %q returned a value from the wrong enum; expected %q", name, expected.Name()))
	}
	return value.ImmutableCopy()
}

func checkedComputedEnumSlice(name string, expected enum.Enum, value enum.ImmutableEnumSlice) enum.ImmutableEnumSlice {
	if value == nil || value.Enum() != expected {
		panic(fmt.Sprintf("computed property %q returned a slice from the wrong enum; expected %q", name, expected.Name()))
	}
	return value.ImmutableCopy()
}

var frameworkComputedProperties = map[string]ComputedPropertyDescriptor{
	"PlayerOrder":       {Scope: ComputedPropertyScopeGlobal, Type: TypePlayerIndexSlice},
	"AvailableTeams":    {Scope: ComputedPropertyScopeGlobal, Type: TypeIllegal},
	"AvailableRoles":    {Scope: ComputedPropertyScopeGlobal, Type: TypeIllegal},
	"AvailableColors":   {Scope: ComputedPropertyScopeGlobal, Type: TypeIllegal},
	"ReadyToStartError": {Scope: ComputedPropertyScopeGlobal, Type: TypeString},
	"Color":             {Scope: ComputedPropertyScopePlayer, Type: TypeString},
	"MayBeActive":       {Scope: ComputedPropertyScopePlayer, Type: TypeBool},
	"GameScore":         {Scope: ComputedPropertyScopePlayer, Type: TypeInt},
	"TeamValue":         {Scope: ComputedPropertyScopePlayer, Type: TypeString},
	"RoleValue":         {Scope: ComputedPropertyScopePlayer, Type: TypeString},
	"ColorValue":        {Scope: ComputedPropertyScopePlayer, Type: TypeString},
	"IsGameAdmin":       {Scope: ComputedPropertyScopePlayer, Type: TypeBool},
}

func (g *GameManager) installComputedProperties(properties []ComputedProperty) error {
	seen := make(map[string]bool)
	for i := range properties {
		property := properties[i]
		if property.configErr != nil {
			return fmt.Errorf("computed property %d is invalid: %v", i, property.configErr)
		}
		name := property.descriptor.Name
		if name == "" {
			return fmt.Errorf("computed property %d has an empty name", i)
		}
		if framework, ok := frameworkComputedProperties[name]; ok &&
			(framework.Scope != property.descriptor.Scope || framework.Type != property.descriptor.Type) {
			return fmt.Errorf("computed property %q conflicts with the framework's %s %s contract", name, framework.Scope, framework.Type)
		}
		key := string(property.descriptor.Scope) + "\x00" + name
		if seen[key] {
			return fmt.Errorf("%s computed property %q is configured more than once", property.descriptor.Scope, name)
		}
		seen[key] = true
		if property.descriptor.Type == TypeEnum || property.descriptor.Type == TypeEnumSlice {
			e := g.chest.Enums().Enum(property.descriptor.EnumName)
			if e == nil {
				return fmt.Errorf("computed property %q references enum %q, which is not in the game chest", name, property.descriptor.EnumName)
			}
			if e != property.configuredEnum {
				return fmt.Errorf("computed property %q uses a different enum instance than chest enum %q", name, property.descriptor.EnumName)
			}
		}
	}
	g.computedProperties = append([]ComputedProperty(nil), properties...)
	return nil
}

// ComputedPropertyDescriptors returns a deterministic copy of the game's
// configured computed-value schema for generators and diagnostics.
func (g *GameManager) ComputedPropertyDescriptors() []ComputedPropertyDescriptor {
	result := make([]ComputedPropertyDescriptor, len(g.computedProperties))
	for i, property := range g.computedProperties {
		result[i] = property.descriptor
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Scope != result[j].Scope {
			return result[i].Scope < result[j].Scope
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func (g *GameManager) computedGlobalProperties(state ImmutableState) PropertyCollection {
	result := g.delegate.FrameworkComputedGlobalProperties(state)
	if result == nil {
		result = PropertyCollection{}
	}
	for _, property := range g.computedProperties {
		if property.descriptor.Scope == ComputedPropertyScopeGlobal {
			result[property.descriptor.Name] = property.global(state)
		}
	}
	return result
}

func (g *GameManager) computedPlayerProperties(player ImmutableSubState) PropertyCollection {
	result := g.delegate.FrameworkComputedPlayerProperties(player)
	if result == nil {
		result = PropertyCollection{}
	}
	for _, property := range g.computedProperties {
		if property.descriptor.Scope == ComputedPropertyScopePlayer {
			result[property.descriptor.Name] = property.player(player)
		}
	}
	return result
}
