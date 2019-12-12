package codegen

import (
	"errors"
	"go/build"
	"go/format"
	"log"
	"strings"
	"unicode"

	"github.com/MarcGrol/golangAnnotations/model"
	"github.com/MarcGrol/golangAnnotations/parser"
	"github.com/jkomoros/boardgame"
)

type memoizedEmbeddedStructKey struct {
	Import           string
	TargetStructName string
}

var memoizedEmbeddedStructs map[memoizedEmbeddedStructKey]fieldsInfo

//fieldsInfo is a collection of field names in a struct and the info about each
//field.
type fieldsInfo map[string]*fieldInfo

//fieldInfo is a collection of information about the specific field in a given
//struct, including their boardgame.PropertyType, whether they're the mutable
//version, and if they're actually a higher type in the hierarchy (e.g.
//MergedStack instead of just a STack).
type fieldInfo struct {
	//Types is the type of the given named field
	Type    boardgame.PropertyType
	Mutable bool
	//SubType will be non-"" if the type is actually represented by a literal
	//value of one of this type's SubTypes, for example a "SizedStack" when the
	//Type is boardgame.TypeStack.
	SubType string
}

//nameForTypeInfo represents each named struct field of a given
//boardgame.PropertyType that we'll output, including enough information to know
//if it needs mutable/upconverter.
type nameForTypeInfo struct {
	Name    string
	Mutable bool
	SubType string
}

const magicDocLinePrefix = "boardgame:codegen"

func init() {
	memoizedEmbeddedStructs = make(map[memoizedEmbeddedStructKey]fieldsInfo)
}

/*

ProcessReaders operates on the package at the given relative location, and
produces two strings, one that is appropriate to be saved in auto_reader.go,
and one that is appropriate to be saved in auto_reader_test.go.

ProcessReaders processes a package of go files, searching for structs that
have a comment immediately above their declaration that begins with
"boardgame:codegen". For each such struct, it creates a Reader(), ReadSetter(),
and ReadSetConfigurer() method that implement boardgame.Reader,
boardgame.ReadSetter, and boardgame.ReadSetConfigurer, respectively.

Producing a ReadSetConfigurator requires a ReadSetter, and producing a
ReadSetter requires a Reader. By default if you have the magic comment of
`boardgame:codegen` it with produce all three. However, if you want only some of
the methods, include an argument for the highest one you want, e.g.
`boardgame:codegen readsetter` to generate a Reader() and ReadSetter().

This package will automatically create additional type transform methods
to handle fields whose literal type is boardgame.ImmutableSizedStack,
boardgame.SizedStack, boardgame.MergedStack, enum.RangeValue, and
enum.TreeValue.

The outputted readers, readsetters, and readsetconfigurers use a hard-
coded list of fields for performance (reflection would be about 30% slower
under normal usage). You should re-generate output every time you add a
struct or modify the fields on a struct.

*/
func ProcessReaders(location string) (output string, testOutput string, err error) {

	sources, err := parser.ParseSourceDir(location, ".*")

	if err != nil {
		return "", "", errors.New("Couldn't parse sources: " + err.Error())
	}

	output, err = doProcessStructs(sources, location, false)

	if err != nil {
		return "", "", errors.New("Couldn't process non-test files: " + err.Error())
	}

	testOutput, err = doProcessStructs(sources, location, true)

	if err != nil {
		return "", "", errors.New("Couldn't process test files: " + err.Error())
	}

	formattedBytes, err := format.Source([]byte(output))

	if err != nil {
		if debugSaveBadCode {
			formattedBytes = []byte(output)
		} else {
			return "", "", errors.New("Couldn't go fmt code for reader: " + err.Error())
		}
	}

	formattedTestBytes, err := format.Source([]byte(testOutput))

	if err != nil {
		if debugSaveBadCode {
			formattedTestBytes = []byte(testOutput)
		} else {
			return "", "", errors.New("Couldn't go fmt code for test reader: " + err.Error())
		}
	}

	return string(formattedBytes), string(formattedTestBytes), nil
}

func doProcessStructs(sources model.ParsedSources, location string, testFiles bool) (output string, err error) {

	if len(sources.Structs) == 0 {
		//If there are no structs at all that's OK, just don't ouput anything.
		return "", nil
	}

	for _, theStruct := range sources.Structs {

		//Only process structs in test or not test files, depending on which mode we're in.
		structInTestFile := strings.HasSuffix(theStruct.Filename, "_test.go")

		if structInTestFile != testFiles {
			continue
		}

		generator := newReaderGenerator(theStruct, location, sources.Structs)

		if generator == nil {
			//No utput necessary
			continue
		}

		output += generator.Output()
	}

	if output != "" {
		output = headerForPackage(sources.Structs[0].PackageName) + output
	}

	return output, nil
}

func fieldNamePublic(name string) bool {
	if len(name) < 1 {
		return false
	}

	firstChar := []rune(name)[0]

	if firstChar != unicode.ToUpper(firstChar) {
		//It was not upper case, thus private, thus should not be included.
		return false
	}

	//TODO: check if the struct says propertyreader:omit

	return true
}

//fieldNamePossibleEmbeddedStruct returns true if it's possible that the field
//is an embedded struct.
func fieldNamePossibleEmbeddedStruct(theField model.Field) bool {

	theTypeParts := strings.Split(theField.TypeName, ".")

	if len(theTypeParts) != 2 {
		return false
	}

	if theField.Name == theTypeParts[0] {
		return true
	}

	return false
}

func (f fieldsInfo) setType(fieldName string, t boardgame.PropertyType) {
	info := f[fieldName]
	if info == nil {
		info = new(fieldInfo)
		f[fieldName] = info
	}
	info.Type = t
}

func (f fieldsInfo) setMutable(fieldName string, isMutable bool) {
	info := f[fieldName]
	if info == nil {
		info = new(fieldInfo)
		f[fieldName] = info
	}
	info.Mutable = isMutable
}

func (f fieldsInfo) setSubType(fieldName string, subType string) {
	info := f[fieldName]
	if info == nil {
		info = new(fieldInfo)
		f[fieldName] = info
	}
	info.SubType = subType
}

func (f fieldsInfo) combine(other fieldsInfo) {
	if other == nil {
		return
	}
	for key, val := range other {
		f[key] = val
	}
}

func structFields(location string, theStruct model.Struct, allStructs []model.Struct) fieldsInfo {

	result := make(fieldsInfo)

	for _, field := range theStruct.Fields {
		if fieldNamePossibleEmbeddedStruct(field) {
			embeddedInfo := typesForPossibleEmbeddedStruct(location, field, allStructs)
			if embeddedInfo != nil {
				result.combine(embeddedInfo)
				continue
			}
		}
		//Check if it's a local-to-package anonymous embedded struct
		if field.Name == "" {
			foundStruct := false
			for _, otherStruct := range allStructs {
				if otherStruct.Name == field.TypeName {
					embeddedInfo := structFields(location, otherStruct, allStructs)
					result.combine(embeddedInfo)
					foundStruct = true
					break
				}
			}
			if foundStruct {
				continue
			}
		}
		if !fieldNamePublic(field.Name) {
			continue
		}
		switch field.TypeName {
		case "int":
			if field.IsSlice {
				result.setType(field.Name, boardgame.TypeIntSlice)
			} else {
				result.setType(field.Name, boardgame.TypeInt)
			}
			result.setMutable(field.Name, true)
		case "bool":
			if field.IsSlice {
				result.setType(field.Name, boardgame.TypeBoolSlice)
			} else {
				result.setType(field.Name, boardgame.TypeBool)
			}
			result.setMutable(field.Name, true)
		case "string":
			if field.IsSlice {
				result.setType(field.Name, boardgame.TypeStringSlice)
			} else {
				result.setType(field.Name, boardgame.TypeString)
			}
			result.setMutable(field.Name, true)
		case "boardgame.PlayerIndex":
			if field.IsSlice {
				result.setType(field.Name, boardgame.TypePlayerIndexSlice)
			} else {
				result.setType(field.Name, boardgame.TypePlayerIndex)
			}
			result.setMutable(field.Name, true)
		case "boardgame.ImmutableStack":
			result.setType(field.Name, boardgame.TypeStack)
			result.setMutable(field.Name, false)
		case "boardgame.MergedStack":
			result.setType(field.Name, boardgame.TypeStack)
			result.setMutable(field.Name, false)
			result.setSubType(field.Name, "MergedStack")
		case "boardgame.Stack":
			result.setType(field.Name, boardgame.TypeStack)
			result.setMutable(field.Name, true)
		case "boardgame.ImmutableSizedStack":
			result.setType(field.Name, boardgame.TypeStack)
			result.setMutable(field.Name, false)
			result.setSubType(field.Name, "ImmutableSizedStack")
		case "boardgame.SizedStack":
			result.setType(field.Name, boardgame.TypeStack)
			result.setMutable(field.Name, true)
			result.setSubType(field.Name, "SizedStack")
		case "boardgame.ImmutableBoard":
			result.setType(field.Name, boardgame.TypeBoard)
			result.setMutable(field.Name, false)
		case "boardgame.Board":
			result.setType(field.Name, boardgame.TypeBoard)
			result.setMutable(field.Name, true)
		case "enum.ImmutableVal":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, false)
		case "enum.Val":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, true)
		case "enum.ImmutableRangeVal":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, false)
			result.setSubType(field.Name, "ImmutableRangeVal")
		case "enum.RangeVal":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, true)
			result.setSubType(field.Name, "RangeVal")
		case "enum.ImmutableTreeVal":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, false)
			result.setSubType(field.Name, "ImmutableTreeVal")
		case "enum.TreeVal":
			result.setType(field.Name, boardgame.TypeEnum)
			result.setMutable(field.Name, true)
			result.setSubType(field.Name, "TreeVal")
		case "boardgame.ImmutableTimer":
			result.setType(field.Name, boardgame.TypeTimer)
			result.setMutable(field.Name, false)
		case "boardgame.Timer":
			result.setType(field.Name, boardgame.TypeTimer)
			result.setMutable(field.Name, true)
		default:
			log.Println("Unknown type on " + theStruct.Name + ": " + field.Name + ": " + field.TypeName)
		}
	}

	return result
}

//typeforPossibleEmbeddedStruct should be called when we think that an unknown
//field MIGHT be an embedded struct. If it is, we will identify the package it
//appears to be built from, parse those structs, try to find the struct, and
//return a map of property types in it.
func typesForPossibleEmbeddedStruct(location string, theField model.Field, allStructs []model.Struct) fieldsInfo {

	targetTypeParts := strings.Split(theField.TypeName, ".")

	if len(targetTypeParts) != 2 {
		return nil
	}

	targetType := targetTypeParts[1]

	//base.SubState will be anonymously embedded but should be ignored.
	if targetTypeParts[0] == "base" && targetType == "SubState" {
		return nil
	}

	key := memoizedEmbeddedStructKey{
		Import:           theField.PackageName,
		TargetStructName: targetType,
	}

	foundTypes := memoizedEmbeddedStructs[key]

	if foundTypes != nil {
		return foundTypes
	}

	pkg, err := build.Import(theField.PackageName, location, build.FindOnly)

	if err != nil {
		log.Println("Couldn't find canonical import: " + err.Error())
		return nil
	}

	importPath := pkg.Dir

	sources, err := parser.ParseSourceDir(importPath, ".*")

	if err != nil {
		log.Println("Error in sources for ", theField, err.Error())
		return nil
	}

	for _, theStruct := range sources.Structs {
		if theStruct.Name != targetType {
			continue
		}
		//Found it!
		foundTypes = structFields(location, theStruct, allStructs)

		memoizedEmbeddedStructs[key] = foundTypes

	}

	return foundTypes

}

func structConfig(docLines []string) (outputReader bool, outputReadSetter bool, outputReadSetConfigurer bool) {

	for _, mainDocLine := range docLines {
		//Multi-line comments will come in as one docline.
		for _, docLine := range strings.Split(mainDocLine, "\n") {
			docLine = strings.ToLower(docLine)
			docLine = strings.TrimPrefix(docLine, "//")
			docLine = strings.TrimSpace(docLine)
			if !strings.HasPrefix(docLine, magicDocLinePrefix) {
				continue
			}
			docLine = strings.TrimPrefix(docLine, magicDocLinePrefix)
			docLine = strings.TrimSpace(docLine)

			switch docLine {
			case "", "all", "readsetconfigurer":
				return true, true, true
			case "readsetter":
				return true, true, false
			case "reader":
				return true, false, false
			}
		}
	}
	return false, false, false
}

func headerForPackage(packageName string) string {

	return templateOutput(headerTemplate, map[string]string{
		"packageName": packageName,
	}) + importText
}
