package codegen

import (
	"errors"
	"github.com/MarcGrol/golangAnnotations/model"
	"github.com/MarcGrol/golangAnnotations/parser"
	"github.com/jkomoros/boardgame"
	"go/build"
	"log"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

var headerTemplate *template.Template
var structHeaderTemplate *template.Template
var typedPropertyTemplate *template.Template
var readerTemplate *template.Template
var readSetterTemplate *template.Template
var readSetConfigurerTemplate *template.Template

type memoizedEmbeddedStructKey struct {
	Import           string
	TargetStructName string
}

var memoizedEmbeddedStructs map[memoizedEmbeddedStructKey]*typeInfo

type typeInfo struct {
	Types       map[string]boardgame.PropertyType
	Mutable     map[string]bool
	UpConverter map[string]string
}

const magicDocLinePrefix = "+autoreader"

func init() {

	funcMap := template.FuncMap{
		"withimmutable": withImmutable,
		"ismutable":     isMutable,
		"verbfortype":   verbForType,
	}

	headerTemplate = template.Must(template.New("header").Funcs(funcMap).Parse(headerTemplateText))
	structHeaderTemplate = template.Must(template.New("structHeader").Funcs(funcMap).Parse(structHeaderTemplateText))
	typedPropertyTemplate = template.Must(template.New("typedProperty").Funcs(funcMap).Parse(typedPropertyTemplateText))
	readerTemplate = template.Must(template.New("reader").Funcs(funcMap).Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Funcs(funcMap).Parse(readSetterTemplateText))
	readSetConfigurerTemplate = template.Must(template.New("readsetconfigurer").Funcs(funcMap).Parse(readSetConfigurerTemplateText))

	memoizedEmbeddedStructs = make(map[memoizedEmbeddedStructKey]*typeInfo)
}

/*

ProcessStructs operates on the package at the given relative location, and
produces two strings, one that is appropriate to be saved in auto_reader.go,
and one that is appropriate to be saved in auto_reader_test.go.

Autoreader processes a package of go files, searching for structs that
have a comment immediately above their declaration that begins with
"+autoreader". For each such struct, it creates a Reader(), ReadSetter(),
and ReadSetConfigurer() method that implement boardgame.Reader,
boardgame.ReadSetter, and boardgame.ReadSetConfigurer, respectively.

Producing a ReadSetConfigurator requires a ReadSetter, and producing a
ReadSetter requires a Reader. By default if you have the magic comment of
`+autoreader` it with produce all three. However, if you want only some of
the methods, include an argument for the highest one you want, e.g.
`+autoreader readsetter` to generate a Reader() and ReadSetter().

This package will automatically create additional type transform methods
to handle fields whose literal type is boardgame.ImmutableSizedStack,
boardgame.SizedStack, boardgame.MergedStack, enum.RangeValue, and
enum.TreeValue.

The outputted readers, readsetters, and readsetconfigurers use a hard-
coded list of fields for performance (reflection would be about 30% slower
under normal usage). You should re-generate output every time you add a
struct or modify the fields on a struct.

*/
func ProcessStructs(location string) (output string, testOutput string, err error) {

	sources, err := parser.ParseSourceDir(location, ".*")

	if err != nil {
		return "", "", errors.New("Couldn't parse sources: " + err.Error())
	}

	output, err = doProcessStructs(sources, location, false)

	if err != nil {
		return
	}

	testOutput, err = doProcessStructs(sources, location, true)

	return
}

func doProcessStructs(sources model.ParsedSources, location string, testFiles bool) (output string, err error) {
	haveOutputHeader := false

	for _, theStruct := range sources.Structs {

		//Only process structs in test or not test files, depending on which mode we're in.
		structInTestFile := strings.HasSuffix(theStruct.Filename, "_test.go")

		if structInTestFile != testFiles {
			continue
		}

		if !haveOutputHeader {
			output += headerForPackage(theStruct.PackageName)
			haveOutputHeader = true
		}

		outputReader, outputReadSetter, outputReadSetConfigurer := structConfig(theStruct.DocLines)

		if !outputReader && !outputReadSetter && !outputReadSetConfigurer {
			continue
		}

		types := structTypes(location, theStruct, sources.Structs)

		output += headerForStruct(theStruct.Name, types, outputReadSetter, outputReadSetConfigurer)

		if outputReader {
			output += readerForStruct(theStruct.Name)
		}
		if outputReadSetter {
			output += readSetterForStruct(theStruct.Name)
		}
		if outputReadSetConfigurer {
			output += readSetConfigurerForStruct(theStruct.Name)
		}
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

func structTypes(location string, theStruct model.Struct, allStructs []model.Struct) *typeInfo {

	result := &typeInfo{
		make(map[string]boardgame.PropertyType),
		make(map[string]bool),
		make(map[string]string),
	}

	for _, field := range theStruct.Fields {
		if fieldNamePossibleEmbeddedStruct(field) {
			embeddedInfo := typesForPossibleEmbeddedStruct(location, field, allStructs)
			if embeddedInfo != nil {
				for key, val := range embeddedInfo.Types {
					result.Types[key] = val
				}
				for key, val := range embeddedInfo.Mutable {
					result.Mutable[key] = val
				}
				continue
			}
		}
		//Check if it's a local-to-package anonymous embedded struct
		if field.Name == "" {
			foundStruct := false
			for _, otherStruct := range allStructs {
				if otherStruct.Name == field.TypeName {
					embeddedInfo := structTypes(location, otherStruct, allStructs)
					for key, val := range embeddedInfo.Types {
						result.Types[key] = val
					}
					for key, val := range embeddedInfo.Mutable {
						result.Mutable[key] = val
					}
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
				result.Types[field.Name] = boardgame.TypeIntSlice
			} else {
				result.Types[field.Name] = boardgame.TypeInt
			}
			result.Mutable[field.Name] = true
		case "bool":
			if field.IsSlice {
				result.Types[field.Name] = boardgame.TypeBoolSlice
			} else {
				result.Types[field.Name] = boardgame.TypeBool
			}
			result.Mutable[field.Name] = true
		case "string":
			if field.IsSlice {
				result.Types[field.Name] = boardgame.TypeStringSlice
			} else {
				result.Types[field.Name] = boardgame.TypeString
			}
			result.Mutable[field.Name] = true
		case "boardgame.PlayerIndex":
			if field.IsSlice {
				result.Types[field.Name] = boardgame.TypePlayerIndexSlice
			} else {
				result.Types[field.Name] = boardgame.TypePlayerIndex
			}
			result.Mutable[field.Name] = true
		case "boardgame.ImmutableStack":
			result.Types[field.Name] = boardgame.TypeStack
			result.Mutable[field.Name] = false
		case "boardgame.MergedStack":
			result.Types[field.Name] = boardgame.TypeStack
			result.Mutable[field.Name] = false
			result.UpConverter[field.Name] = "MergedStack"
		case "boardgame.Stack":
			result.Types[field.Name] = boardgame.TypeStack
			result.Mutable[field.Name] = true
		case "boardgame.ImmutableSizedStack":
			result.Types[field.Name] = boardgame.TypeStack
			result.Mutable[field.Name] = false
			result.UpConverter[field.Name] = "ImmutableSizedStack"
		case "boardgame.SizedStack":
			result.Types[field.Name] = boardgame.TypeStack
			result.Mutable[field.Name] = true
			result.UpConverter[field.Name] = "SizedStack"
		case "boardgame.ImmutableBoard":
			result.Types[field.Name] = boardgame.TypeBoard
			result.Mutable[field.Name] = false
		case "boardgame.Board":
			result.Types[field.Name] = boardgame.TypeBoard
			result.Mutable[field.Name] = true
		case "enum.ImmutableVal":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = false
		case "enum.Val":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = true
		case "enum.ImmutableRangeVal":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = false
			result.UpConverter[field.Name] = "ImmutableRangeVal"
		case "enum.RangeVal":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = true
			result.UpConverter[field.Name] = "RangeVal"
		case "enum.ImmutableTreeVal":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = false
			result.UpConverter[field.Name] = "ImmutableTreeVal"
		case "enum.TreeVal":
			result.Types[field.Name] = boardgame.TypeEnum
			result.Mutable[field.Name] = true
			result.UpConverter[field.Name] = "TreeVal"
		case "boardgame.ImmutableTimer":
			result.Types[field.Name] = boardgame.TypeTimer
			result.Mutable[field.Name] = false
		case "boardgame.Timer":
			result.Types[field.Name] = boardgame.TypeTimer
			result.Mutable[field.Name] = true
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
func typesForPossibleEmbeddedStruct(location string, theField model.Field, allStructs []model.Struct) *typeInfo {

	targetTypeParts := strings.Split(theField.TypeName, ".")

	if len(targetTypeParts) != 2 {
		return nil
	}

	targetType := targetTypeParts[1]

	//BaseSubState will be anonymously embedded but should be ignored.
	if targetType == "BaseSubState" {
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
		foundTypes = structTypes(location, theStruct, allStructs)

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

func withImmutable(in string) string {
	prefix := ""
	rest := in
	parts := strings.Split(in, ".")
	if len(parts) > 1 {
		prefix = strings.Join(parts[:len(parts)-1], ".")
		rest = parts[len(parts)-1]
	}

	if _, needsImmutable := configureTypes[rest]; !needsImmutable {
		return in
	}

	rest = "Immutable" + rest

	if prefix == "" {
		return rest
	}

	return prefix + "." + rest

}

func isMutable(in string) bool {

	for key := range configureTypes {
		if strings.Contains(in, key) {
			return true
		}
	}

	return false
}

var configureTypes = map[string]bool{
	"Stack": true,
	"Timer": true,
	"Board": true,
	"Val":   true,
	"Enum":  true,
}

func verbForType(in string) string {
	_, configure := configureTypes[in]
	if configure {
		return "Configure"
	}
	return "Set"
}

type nameForTypeInfo struct {
	Name        string
	Mutable     bool
	UpConverter string
}

func headerForStruct(structName string, types *typeInfo, outputReadSetter bool, outputReadSetConfigurer bool) string {

	//TODO: memoize propertyTypes/setterPropertyTypes because they don't
	//change within a run of this program.

	//propertyTypes is short name, golangValue
	propertyTypes := make(map[string]string)
	setterPropertyTypes := make(map[string]string)

	//readSetConfigurer is a superset of readSetter, which means that if
	//output readSetConfigurer we must also output readSetter.
	if outputReadSetConfigurer {
		outputReadSetter = true
	}

	for i := boardgame.TypeInt; i <= boardgame.TypeTimer; i++ {

		key := strings.TrimPrefix(i.String(), "Type")

		setterKey := ""

		goLangType := key
		setterGoLangType := ""
		switch key {
		case "Bool":
			goLangType = "bool"
		case "Int":
			goLangType = "int"
		case "String":
			goLangType = "string"
		case "PlayerIndex":
			goLangType = "boardgame.PlayerIndex"
		case "IntSlice":
			goLangType = "[]int"
		case "BoolSlice":
			goLangType = "[]bool"
		case "StringSlice":
			goLangType = "[]string"
		case "PlayerIndexSlice":
			goLangType = "[]boardgame.PlayerIndex"
		case "Enum":
			goLangType = "enum.ImmutableVal"
			setterKey = "Enum"
			setterGoLangType = "enum.Val"
		case "Stack":
			goLangType = "boardgame.ImmutableStack"
			setterKey = "Stack"
			setterGoLangType = "boardgame.Stack"
		case "Board":
			goLangType = "boardgame.ImmutableBoard"
			setterKey = "Board"
			setterGoLangType = "boardgame.Board"
		case "Timer":
			goLangType = "boardgame.ImmutableTimer"
			setterKey = "Timer"
			setterGoLangType = "boardgame.Timer"
		default:
			goLangType = "UNKNOWN"
		}

		if setterKey == "" {
			setterKey = key
		}

		if setterGoLangType == "" {
			setterGoLangType = goLangType
		}

		propertyTypes[key] = goLangType
		setterPropertyTypes[setterKey] = setterGoLangType
	}

	output := templateOutput(structHeaderTemplate, map[string]interface{}{
		"structName":              structName,
		"firstLetter":             strings.ToLower(structName[:1]),
		"readerName":              "__" + structName + "Reader",
		"propertyTypes":           propertyTypes,
		"setterPropertyTypes":     setterPropertyTypes,
		"types":                   types,
		"outputReadSetter":        outputReadSetter,
		"outputReadSetConfigurer": outputReadSetConfigurer,
	})

	sortedKeys := make([]string, len(propertyTypes))
	i := 0

	for propType := range propertyTypes {
		sortedKeys[i] = propType
		i++
	}

	sort.Strings(sortedKeys)

	for _, propType := range sortedKeys {

		goLangType := propertyTypes[propType]

		zeroValue := "nil"

		switch propType {
		case "Bool":
			zeroValue = "false"
		case "Int":
			zeroValue = "0"
		case "String":
			zeroValue = "\"\""
		case "PlayerIndex":
			zeroValue = "0"
		case "IntSlice":
			zeroValue = "[]int{}"
		case "BoolSlice":
			zeroValue = "[]bool{}"
		case "StringSlice":
			zeroValue = "[]string{}"
		case "PlayerIndexSlice":
			zeroValue = "[]boardgame.PlayerIndex{}"
		}

		var namesForType []nameForTypeInfo

		for key, val := range types.Types {
			if val.String() == "Type"+propType {
				namesForType = append(namesForType, nameForTypeInfo{
					Name:        key,
					Mutable:     types.Mutable[key],
					UpConverter: types.UpConverter[key],
				})
			}
		}

		sort.Slice(namesForType, func(i, j int) bool {
			return namesForType[i].Name < namesForType[j].Name
		})

		setterPropType := propType

		outputMutableGetter := false

		switch propType {
		case "Enum":
			setterPropType = "Enum"
			outputMutableGetter = true
		case "Stack":
			setterPropType = "Stack"
			outputMutableGetter = true
		case "Board":
			setterPropType = "Board"
			outputMutableGetter = true
		case "Timer":
			setterPropType = "Timer"
			outputMutableGetter = true
		}

		setterGoLangType := setterPropertyTypes[setterPropType]

		output += templateOutput(typedPropertyTemplate, map[string]interface{}{
			"structName":              structName,
			"firstLetter":             strings.ToLower(structName[:1]),
			"readerName":              "__" + structName + "Reader",
			"propType":                propType,
			"setterPropType":          setterPropType,
			"namesForType":            namesForType,
			"goLangType":              goLangType,
			"setterGoLangType":        setterGoLangType,
			"outputMutableGetter":     outputMutableGetter,
			"zeroValue":               zeroValue,
			"outputReadSetter":        outputReadSetter,
			"outputReadSetConfigurer": outputReadSetConfigurer,
		})
	}

	return output

}

func readerForStruct(structName string) string {

	return templateOutput(readerTemplate, map[string]string{
		"firstLetter": strings.ToLower(structName[:1]),
		"structName":  structName,
		"readerName":  "__" + structName + "Reader",
	})

}

func readSetterForStruct(structName string) string {

	return templateOutput(readSetterTemplate, map[string]string{
		"firstLetter": strings.ToLower(structName[:1]),
		"structName":  structName,
		"readerName":  "__" + structName + "Reader",
	})

}

func readSetConfigurerForStruct(structName string) string {

	return templateOutput(readSetConfigurerTemplate, map[string]string{
		"firstLetter": strings.ToLower(structName[:1]),
		"structName":  structName,
		"readerName":  "__" + structName + "Reader",
	})

}

const headerTemplateText = `/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.SubState and boardgame.MutableSubState. It was 
 * generated by autoreader.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package {{.packageName}}
`

const importText = `import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

`

const structHeaderTemplateText = `// Implementation for {{.structName}}

var __{{.structName}}ReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	{{range $key, $value := .types.Types -}}
		"{{$key}}": boardgame.{{$value.String}},
	{{end}}
}

type {{.readerName}} struct {
	data *{{.structName}}
}

func ({{.firstLetter}} *{{.readerName}}) Props() map[string]boardgame.PropertyType {
	return __{{.structName}}ReaderProps
}

func ({{.firstLetter}} *{{.readerName}}) Prop(name string) (interface{}, error) {
	props := {{.firstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	{{$firstLetter := .firstLetter}}

	switch propType {
	{{range $type, $goLangtype := .propertyTypes -}}
	case boardgame.Type{{$type}}:
		return {{$firstLetter}}.{{withimmutable $type}}Prop(name)
	{{end}}
	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

{{if .outputReadSetter -}}

func ({{.firstLetter}} *{{.readerName}}) PropMutable(name string) bool {
	switch name {
		{{range $key, $val := .types.Mutable -}}
	case "{{$key}}":
		return {{$val}}
		{{end -}}
	}

	return false
}

func ({{.firstLetter}} *{{.readerName}}) SetProp(name string, value interface{}) error {
	props := {{.firstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	{{range $type, $goLangType := .setterPropertyTypes -}}
	{{if ismutable $type -}}
	case boardgame.Type{{$type}}:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	{{- else -}}
	case boardgame.Type{{$type}}:
		val, ok := value.({{$goLangType}})
		if !ok {
			return errors.New("Provided value was not of type {{$goLangType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
	{{- end}}
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}

{{if .outputReadSetConfigurer -}}
func ({{.firstLetter}} *{{.readerName}}) ConfigureProp(name string, value interface{}) error {
	props := {{.firstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	{{range $type, $goLangType := .setterPropertyTypes -}}
	case boardgame.Type{{$type}}:
		{{if ismutable $type -}}
		if {{$firstLetter}}.PropMutable(name) {
			//Mutable variant
			val, ok := value.({{$goLangType}})
			if !ok {
				return errors.New("Provided value was not of type {{$goLangType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
		} else {
			//Immutable variant
			val, ok := value.({{withimmutable $goLangType}})
			if !ok {
				return errors.New("Provided value was not of type {{withimmutable $goLangType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{withimmutable $type}}Prop(name, val)
		}
		{{- else -}}
			val, ok := value.({{$goLangType}})
			if !ok {
				return errors.New("Provided value was not of type {{$goLangType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
		{{- end}}
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}
`

const typedPropertyTemplateText = `func ({{.firstLetter}} *{{.readerName}}) {{withimmutable .propType}}Prop(name string) ({{.goLangType}}, error) {
	{{$firstLetter := .firstLetter}}
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.Name}}":
				return {{$firstLetter}}.data.{{.Name}}, nil
		{{end}}
	}
	{{end}}

	return {{.zeroValue}}, errors.New("No such {{.propType}} prop: " + name)

}

{{if .outputReadSetConfigurer -}}
{{if .outputMutableGetter -}}
func ({{.firstLetter}} *{{.readerName}}) Configure{{.setterPropType}}Prop(name string, value {{.setterGoLangType}}) error {
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				{{if .UpConverter -}}
				slotValue := value.{{.UpConverter}}()
				if slotValue == nil {
					return errors.New("{{.Name}} couldn't be upconverted, returned nil.")
				}
				{{$firstLetter}}.data.{{.Name}} = slotValue
				{{- else -}}
				{{$firstLetter}}.data.{{.Name}} = value
				{{- end}}
				return nil
			{{- else -}}
				return boardgame.ErrPropertyImmutable
			{{- end}}
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.setterPropType}} prop: " + name)

}

func ({{.firstLetter}} *{{.readerName}}) Configure{{withimmutable .setterPropType}}Prop(name string, value {{withimmutable .setterGoLangType}}) error {
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return boardgame.ErrPropertyImmutable
			{{- else -}}
				{{if .UpConverter -}}
				slotValue := value.{{.UpConverter}}()
				if slotValue == nil {
					return errors.New("{{.Name}} couldn't be upconverted, returned nil.")
				}
				{{$firstLetter}}.data.{{.Name}} = slotValue
				{{- else -}}
				{{$firstLetter}}.data.{{.Name}} = value
				{{- end}}
				return nil
			{{- end}}
		{{end}}
	}
	{{end}}

	return errors.New("No such {{withimmutable .setterPropType}} prop: " + name)

}

{{end}}
{{end}}

{{if .outputReadSetter -}}
{{if .outputMutableGetter -}}
func ({{.firstLetter}} *{{.readerName}}) {{.setterPropType}}Prop(name string) ({{.setterGoLangType}}, error) {
	{{$firstLetter := .firstLetter}}
	{{$zeroValue := .zeroValue}}
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return {{$firstLetter}}.data.{{.Name}}, nil
			{{- else -}}
				return {{$zeroValue}}, boardgame.ErrPropertyImmutable
			{{- end}}
		{{end}}
	}
	{{end}}

	return {{.zeroValue}}, errors.New("No such {{.propType}} prop: " + name)

}

{{else}}
func ({{.firstLetter}} *{{.readerName}}) Set{{.setterPropType}}Prop(name string, value {{.setterGoLangType}}) error {
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.Name}}":
				{{$firstLetter}}.data.{{.Name}} = value
				return nil
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.setterPropType}} prop: " + name)

}

{{end}}
{{end}}
`

const readerTemplateText = `func ({{.firstLetter}} *{{.structName}}) Reader() boardgame.PropertyReader {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const readSetterTemplateText = `func ({{.firstLetter}} *{{.structName}}) ReadSetter() boardgame.PropertyReadSetter {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const readSetConfigurerTemplateText = `func ({{.firstLetter}} *{{.structName}}) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`
