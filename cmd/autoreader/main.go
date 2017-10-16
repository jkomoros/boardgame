/*

	Autoreader is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter.

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

	You can configure which package to process and where to write output via
	command-line flags. By default it processes the current package and writes
	its output to auto_reader.go, overwriting whatever file was there before.
	See command-line options by passing -h.

	The outputted readers, readsetters, and readsetconfigurers use a hard-
	coded list of fields for performance (reflection would be about 30% slower
	under normal usage). You should re-run go generate every time you add a
	struct or modify the fields on a struct.

	The defaults are set reasonably so that you can use go:generate very
	easily. See examplepkg/ for a very simple example.

*/
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/MarcGrol/golangAnnotations/model"
	"github.com/MarcGrol/golangAnnotations/parser"
	"github.com/jkomoros/boardgame"
	"go/build"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
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

var memoizedEmbeddedStructs map[memoizedEmbeddedStructKey]map[string]boardgame.PropertyType

const magicDocLinePrefix = "+autoreader"

type appOptions struct {
	OutputFile       string
	PackageDirectory string
	PrintToConsole   bool
	Help             bool
	flagSet          *flag.FlagSet
}

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func init() {

	funcMap := template.FuncMap{
		"withoutmutable": withoutMutable,
		"ismutable":      isMutable,
		"verbfortype":    verbForType,
	}

	headerTemplate = template.Must(template.New("header").Funcs(funcMap).Parse(headerTemplateText))
	structHeaderTemplate = template.Must(template.New("structHeader").Funcs(funcMap).Parse(structHeaderTemplateText))
	typedPropertyTemplate = template.Must(template.New("typedProperty").Funcs(funcMap).Parse(typedPropertyTemplateText))
	readerTemplate = template.Must(template.New("reader").Funcs(funcMap).Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Funcs(funcMap).Parse(readSetterTemplateText))
	readSetConfigurerTemplate = template.Must(template.New("readsetconfigurer").Funcs(funcMap).Parse(readSetConfigurerTemplateText))

	memoizedEmbeddedStructs = make(map[memoizedEmbeddedStructKey]map[string]boardgame.PropertyType)
}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.PackageDirectory, "pkg", ".", "Which package to process")
	options.flagSet.BoolVar(&options.Help, "h", false, "If set, print help message and quit.")
	options.flagSet.BoolVar(&options.PrintToConsole, "print", false, "If true, will print result to console instead of writing to out.")
}

func getOptions(flagSet *flag.FlagSet, flagArguments []string) *appOptions {
	options := &appOptions{flagSet: flagSet}
	defineFlags(options)
	flagSet.Parse(flagArguments)
	return options
}

func main() {
	flagSet := flag.CommandLine
	process(getOptions(flagSet, os.Args[1:]), os.Stdout, os.Stderr)
}

func process(options *appOptions, out io.ReadWriter, errOut io.ReadWriter) {

	if options.Help {
		options.flagSet.SetOutput(out)
		options.flagSet.PrintDefaults()
		return
	}

	output, err := processPackage(options.PackageDirectory)

	if err != nil {
		fmt.Fprintln(errOut, "ERROR", err)
		return
	}

	if options.PrintToConsole {
		fmt.Fprintln(out, output)
	} else {
		ioutil.WriteFile(options.OutputFile, []byte(output), 0644)
	}

}

func processPackage(location string) (output string, err error) {

	sources, err := parser.ParseSourceDir(location, ".*")

	if err != nil {
		return "", errors.New("Couldn't parse sources: " + err.Error())
	}

	haveOutputHeader := false

	for _, theStruct := range sources.Structs {

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

	formattedBytes, err := format.Source([]byte(output))

	if err != nil {
		return "", errors.New("Couldn't go fmt code: " + err.Error())
	}

	return string(formattedBytes), nil
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

func structTypes(location string, theStruct model.Struct, allStructs []model.Struct) map[string]boardgame.PropertyType {
	result := make(map[string]boardgame.PropertyType)
	for _, field := range theStruct.Fields {
		if fieldNamePossibleEmbeddedStruct(field) {
			embeddedInfo := typesForPossibleEmbeddedStruct(location, field, allStructs)
			if embeddedInfo != nil {
				for key, val := range embeddedInfo {
					result[key] = val
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
					for key, val := range embeddedInfo {
						result[key] = val
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
				result[field.Name] = boardgame.TypeIntSlice
			} else {
				result[field.Name] = boardgame.TypeInt
			}
		case "bool":
			if field.IsSlice {
				result[field.Name] = boardgame.TypeBoolSlice
			} else {
				result[field.Name] = boardgame.TypeBool
			}
		case "string":
			if field.IsSlice {
				result[field.Name] = boardgame.TypeStringSlice
			} else {
				result[field.Name] = boardgame.TypeString
			}
		case "boardgame.MutableStack":
			result[field.Name] = boardgame.TypeStack
		case "enum.MutableVal":
			result[field.Name] = boardgame.TypeEnum
		case "boardgame.PlayerIndex":
			if field.IsSlice {
				result[field.Name] = boardgame.TypePlayerIndexSlice
			} else {
				result[field.Name] = boardgame.TypePlayerIndex
			}
		case "boardgame.MutableTimer":
			result[field.Name] = boardgame.TypeTimer
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
func typesForPossibleEmbeddedStruct(location string, theField model.Field, allStructs []model.Struct) map[string]boardgame.PropertyType {

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

	result := memoizedEmbeddedStructs[key]

	if result != nil {
		return result
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
		result = structTypes(location, theStruct, allStructs)

		memoizedEmbeddedStructs[key] = result

		return result
	}

	return nil

}

func structConfig(docLines []string) (outputReader bool, outputReadSetter bool, outputReadSetConfigurer bool) {

	for _, docLine := range docLines {
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
	return false, false, false
}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}

func headerForPackage(packageName string) string {

	return templateOutput(headerTemplate, map[string]string{
		"packageName": packageName,
	}) + importText
}

func withoutMutable(in string) string {
	return strings.Replace(in, "Mutable", "", -1)
}

func isMutable(in string) bool {
	return strings.Contains(in, "Mutable")
}

func verbForType(in string) string {
	if strings.HasPrefix(in, "Mutable") {
		return "Configure"
	}
	return "Set"
}

func headerForStruct(structName string, types map[string]boardgame.PropertyType, outputReadSetter bool, outputReadSetConfigurer bool) string {

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
		case "Enum":
			goLangType = "enum.Val"
			setterKey = "MutableEnum"
			setterGoLangType = "enum.MutableVal"
		case "Stack":
			goLangType = "boardgame.Stack"
			setterKey = "MutableStack"
			setterGoLangType = "boardgame.MutableStack"
		case "IntSlice":
			goLangType = "[]int"
		case "BoolSlice":
			goLangType = "[]bool"
		case "StringSlice":
			goLangType = "[]string"
		case "PlayerIndexSlice":
			goLangType = "[]boardgame.PlayerIndex"
		case "Timer":
			goLangType = "boardgame.Timer"
			setterKey = "MutableTimer"
			setterGoLangType = "boardgame.MutableTimer"
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

	for propType, _ := range propertyTypes {
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

		var namesForType []string

		for key, val := range types {
			if val.String() == "Type"+propType {
				namesForType = append(namesForType, key)
			}
		}

		setterPropType := propType

		outputMutableGetter := false

		switch propType {
		case "Enum":
			setterPropType = "MutableEnum"
			outputMutableGetter = true
		case "Stack":
			setterPropType = "MutableStack"
			outputMutableGetter = true
		case "Timer":
			setterPropType = "MutableTimer"
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
	{{range $key, $value := .types -}}
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
		return {{$firstLetter}}.{{$type}}Prop(name)
	{{end}}
	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

{{if .outputReadSetter -}}
func ({{.firstLetter}} *{{.readerName}}) SetProp(name string, value interface{}) error {
	props := {{.firstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	{{range $type, $goLangType := .setterPropertyTypes -}}
	{{if ismutable $type -}}
	case boardgame.Type{{withoutmutable $type}}:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	{{- else -}}
	case boardgame.Type{{withoutmutable $type}}:
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
	case boardgame.Type{{withoutmutable $type}}:
		val, ok := value.({{$goLangType}})
		if !ok {
			return errors.New("Provided value was not of type {{$goLangType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}
`

const typedPropertyTemplateText = `func ({{.firstLetter}} *{{.readerName}}) {{.propType}}Prop(name string) ({{.goLangType}}, error) {
	{{$firstLetter := .firstLetter}}
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.}}":
				return {{$firstLetter}}.data.{{.}}, nil
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
			case "{{.}}":
				{{$firstLetter}}.data.{{.}} = value
				return nil
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.setterPropType}} prop: " + name)

}

{{end}}
{{end}}

{{if .outputReadSetter -}}
{{if .outputMutableGetter -}}
func ({{.firstLetter}} *{{.readerName}}) {{.setterPropType}}Prop(name string) ({{.setterGoLangType}}, error) {
	{{$firstLetter := .firstLetter}}
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.}}":
				return {{$firstLetter}}.data.{{.}}, nil
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
			case "{{.}}":
				{{$firstLetter}}.data.{{.}} = value
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
