package codegen

import (
	"log"
	"strings"

	"github.com/jkomoros/boardgame"
)

//allValidTypes is an enumeration of all types in order.
var allValidTypes []propertyType

//highestProperty is the highest enum in the PropertyType enum.
const highestProperty = boardgame.TypeTimer

func init() {
	//Only need space for highestProperty because we skip TypeIllegal.
	allValidTypes = make([]propertyType, highestProperty)
	//We skip TypeIllegal
	for i := 0; i < int(highestProperty); i++ {
		allValidTypes[i] = propertyType{boardgame.PropertyType(i + 1)}
	}
}

//propertyType is a simple wrapper around boardgame.PropertyType that extends it
//with methods that are only useful for codegen and thus don't belong in the
//main package. We can't just alias the type in this package because we need the
//methods. Note one oddity, to compare what type it is, you have to compare
//t.PropertyType == boardgame.TypeInt. Luckily the typechecker will complain if
//you do it wrong.
type propertyType struct {
	boardgame.PropertyType
}

//TypePackagePrefix returns a string representing the package prefix of the go
//type that is represented by this property type. For example, "boardgame." for
//TypeStack, and "" for TypeInt. Using strings.TrimPrefix() with this prefix
//applied to the return value of for example ImmutableGoType and others will
//strip away the package qualifier, if it exists. Most useful for the codegen
//package.
func (t propertyType) TypePackagePrefix() string {
	//Strip away any slices so we have fewer conditions to test for
	base := t.BaseType()
	switch base {
	case boardgame.TypePlayerIndex, boardgame.TypeStack, boardgame.TypeBoard, boardgame.TypeTimer:
		return "boardgame."
	case boardgame.TypeEnum:
		return "enum."
	}
	return ""
}

//ImmutableGoType emits strings like 'bool', 'boardgame.PlayerIndex'. It
//represents the type of this property for the immutable/getter contexts. Most
//useful for codegen package.
func (t propertyType) ImmutableGoType() string {

	if t.IsSlice() {
		return "[]" + propertyType{t.BaseType()}.ImmutableGoType()
	}

	switch t.PropertyType {
	case boardgame.TypeBool:
		return "bool"
	case boardgame.TypeInt:
		return "int"
	case boardgame.TypeString:
		return "string"
	case boardgame.TypePlayerIndex:
		return "boardgame.PlayerIndex"
	case boardgame.TypeEnum:
		return "enum.ImmutableVal"
	case boardgame.TypeStack:
		return "boardgame.ImmutableStack"
	case boardgame.TypeBoard:
		return "boardgame.ImmutableBoard"
	case boardgame.TypeTimer:
		return "boardgame.ImmutableTimer"
	default:
		return ""
	}
}

//MutableGoType emits a string representing the golang type for the property
//when in mutable/setter contexts, e.g 'int', 'boardgame.Stack'. Most useful for
//the codegen package.
func (t propertyType) MutableGoType() string {
	return strings.Replace(t.ImmutableGoType(), "Immutable", "", -1)
}

//Key returns the part of the PropertyReader method signature for this type. For
//example, "Bool" for TypeBool, "Timer" for "TypeTimer". Most useful for the
//codegen package.
func (t propertyType) Key() string {
	return strings.TrimPrefix(t.String(), "Type")
}

//ImmutableSubTypeConverters will return the names of hte converters to convert
//to the various valid subtypes, e.g. "ImmutableSizedStack" for TypeStack.
func (t propertyType) ImmutableSubTypeConverters() []string {
	switch t.PropertyType {
	case boardgame.TypeStack:
		return []string{"MergedStack", "ImmutableSizedStack"}
	case boardgame.TypeEnum:
		return []string{"ImmutableTreeVal", "ImmutableRangeVal"}
	}
	return nil
}

//MutableSubTypeConverters will return the names of hte converters to convert
//to the various valid subtypes, e.g. "ImmutableSizedStack" for TypeStack.
func (t propertyType) MutableSubTypeConverters() []string {
	switch t.PropertyType {
	case boardgame.TypeStack:
		return []string{"SizedStack"}
	case boardgame.TypeEnum:
		return []string{"TreeVal", "RangeVal"}
	}
	return nil
}

//ImmutableSubTypes returns the sub type fully qualitfied type strings for
//sub-types, e.g. "enum.ImmutableRangeVal" for boardgame.TypeEnum.
func (t propertyType) ImmutableSubTypes() []string {
	//Since the convertesr by convention are literally just the name of the type
	//(minus the package qualifier) we can just strip the package name.
	prefix := t.TypePackagePrefix()
	var result []string
	for _, item := range t.ImmutableSubTypeConverters() {
		result = append(result, strings.TrimPrefix(item, prefix))
	}
	return result
}

//MutableSubTypes returns the sub type fully qualitfied type strings for
//sub-types, e.g. "enum.ImmutableRangeVal" for boardgame.TypeEnum.
func (t propertyType) MutableSubTypes() []string {
	//Since the convertesr by convention are literally just the name of the type
	//(minus the package qualifier) we can just strip the package name.
	prefix := t.TypePackagePrefix()
	var result []string
	for _, item := range t.MutableSubTypeConverters() {
		result = append(result, strings.TrimPrefix(item, prefix))
	}
	return result
}

//ZeroValue returns the string representing the zeroValue for this type, e.g.
//"0" for TypeInt and "[]boardgame.PlayerIndex{}" for TypePlayerIndexSlice. Most
//useful for codgen package.
func (t propertyType) ZeroValue() string {

	switch t.PropertyType {
	case boardgame.TypeBool:
		return "false"
	case boardgame.TypeInt:
		return "0"
	case boardgame.TypeString:
		return "\"\""
	case boardgame.TypePlayerIndex:
		return "0"
	case boardgame.TypeIllegal:
		return ""
	}

	if t.IsSlice() {
		return t.ImmutableGoType() + "{}"
	}
	if t.IsInterface() {
		return "nil"
	}

	log.Println("Unexpected type for ZeroValue")
	return ""

}
