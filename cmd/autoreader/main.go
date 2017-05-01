/*

	Autoreader is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter.

	Autoreader processes a package of go files, searching for structs that
	have a comment immediately above their declaration that begins with
	"+autoreader". For each such struct, it creates a Reader() and
	PropertyReader() method that just use boardgame.DefaultReader and
	boardgame.DefaultReadSetter.

	If you want only a reader or only a readsetter for a given struct, include
	the keyword "reader" or "readsetter", like so: "+autoreader reader"

	You can configure which package to process and where to write output via
	command-line flags. By default it processes the current package and writes
	its output to auto_reader.go, overwriting whatever file was there before.
	See command-line options by passing -h.

	By default readers and readSetters that rely on hard-coded lists of
	properties will be generated. These are faster to execute because they
	don't rely on reflection. However, every time you add or change properties
	to a struct, you must re-run go generate. Another option is available that
	uses reflection (via board.DefaultReader) to implement the Readers and
	ReadSetters (by passing -reflect). The pro is that you only need to run
	`go generate` when you add or remove a struct; the downside is that run-
	time performance will be worse (roughly 30% worse in typical workloads).

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
var reflectStructHeaderTemplate *template.Template
var readerTemplate *template.Template
var reflectReaderTemplate *template.Template
var readSetterTemplate *template.Template
var reflectReadSetterTemplate *template.Template

const magicDocLinePrefix = "+autoreader"

type appOptions struct {
	OutputFile       string
	PackageDirectory string
	PrintToConsole   bool
	Help             bool
	UseReflection    bool
	flagSet          *flag.FlagSet
}

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func init() {
	headerTemplate = template.Must(template.New("header").Parse(headerTemplateText))
	structHeaderTemplate = template.Must(template.New("structHeader").Parse(structHeaderTemplateText))
	typedPropertyTemplate = template.Must(template.New("typedProperty").Parse(typedPropertyTemplateText))
	reflectStructHeaderTemplate = template.Must(template.New("structHeader").Parse(reflectStructHeaderTemplateText))
	readerTemplate = template.Must(template.New("reader").Parse(readerTemplateText))
	reflectReaderTemplate = template.Must(template.New("reader").Parse(reflectReaderTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Parse(readSetterTemplateText))
	reflectReadSetterTemplate = template.Must(template.New("readsetter").Parse(reflectReadSetterTemplateText))

}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.PackageDirectory, "pkg", ".", "Which package to process")
	options.flagSet.BoolVar(&options.Help, "h", false, "If set, print help message and quit.")
	options.flagSet.BoolVar(&options.PrintToConsole, "print", false, "If true, will print result to console instead of writing to out.")
	options.flagSet.BoolVar(&options.UseReflection, "reflect", false, "If true, will use reflection based output.")
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

	output, err := processPackage(options.UseReflection, options.PackageDirectory)

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

func processPackage(useReflection bool, location string) (output string, err error) {
	sources, err := parser.ParseSourceDir(location, ".*")

	if err != nil {
		return "", errors.New("Couldn't parse sources: " + err.Error())
	}

	haveOutputHeader := false

	for _, theStruct := range sources.Structs {

		if !haveOutputHeader {
			output += headerForPackage(useReflection, theStruct.PackageName)
			haveOutputHeader = true
		}

		outputReader, outputReadSetter := structConfig(theStruct.DocLines)

		types := structTypes(theStruct)

		if outputReader || outputReadSetter {
			output += headerForStruct(useReflection, theStruct.Name, types, outputReadSetter)
		}

		if outputReader {
			output += readerForStruct(useReflection, theStruct.Name)
		}
		if outputReadSetter {
			output += readSetterForStruct(useReflection, theStruct.Name)
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

func structTypes(theStruct model.Struct) map[string]boardgame.PropertyType {
	result := make(map[string]boardgame.PropertyType)
	for _, field := range theStruct.Fields {
		if !fieldNamePublic(field.Name) {
			continue
		}
		switch field.TypeName {
		case "int":
			result[field.Name] = boardgame.TypeInt
		case "bool":
			result[field.Name] = boardgame.TypeBool
		case "string":
			result[field.Name] = boardgame.TypeString
		case "boardgame.SizedStack":
			result[field.Name] = boardgame.TypeSizedStack
		case "boardgame.GrowableStack":
			result[field.Name] = boardgame.TypeGrowableStack
		case "boardgame.PlayerIndex":
			result[field.Name] = boardgame.TypePlayerIndex
		case "boardgame.Timer":
			result[field.Name] = boardgame.TypeTimer
		default:
			log.Println("Unknown type on " + theStruct.Name + ": " + field.Name + ": " + field.TypeName)
		}
	}
	return result
}

func structConfig(docLines []string) (outputReader bool, outputReadSetter bool) {

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
		case "":
			return true, true
		case "both":
			return true, true
		case "reader":
			return true, false
		case "readsetter":
			return false, true
		}

	}
	return false, false
}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}

func headerForPackage(useReflection bool, packageName string) string {

	importTextToUse := importText

	if useReflection {
		importTextToUse = reflectImportText
	}

	return templateOutput(headerTemplate, map[string]string{
		"packageName": packageName,
	}) + importTextToUse
}

func headerForStruct(useReflection bool, structName string, types map[string]boardgame.PropertyType, outputReadSetter bool) string {

	if useReflection {
		return templateOutput(reflectStructHeaderTemplate, map[string]string{
			"structName": structName,
		})
	}

	//propertyTypes is short name, golangValue
	propertyTypes := make(map[string]string)

	for i := boardgame.TypeInt; i <= boardgame.TypeTimer; i++ {

		key := strings.TrimPrefix(i.String(), "Type")

		goLangType := key
		switch key {
		case "Bool":
			goLangType = "bool"
		case "Int":
			goLangType = "int"
		case "String":
			goLangType = "string"
		case "PlayerIndex":
			goLangType = "boardgame.PlayerIndex"
		default:
			goLangType = "*boardgame." + goLangType
		}

		propertyTypes[key] = goLangType
	}

	output := templateOutput(structHeaderTemplate, map[string]interface{}{
		"structName":       structName,
		"firstLetter":      strings.ToLower(structName[:1]),
		"readerName":       "__" + structName + "Reader",
		"propertyTypes":    propertyTypes,
		"types":            types,
		"outputReadSetter": outputReadSetter,
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
		}

		var namesForType []string

		for key, val := range types {
			if val.String() == "Type"+propType {
				namesForType = append(namesForType, key)
			}
		}

		output += templateOutput(typedPropertyTemplate, map[string]interface{}{
			"structName":       structName,
			"firstLetter":      strings.ToLower(structName[:1]),
			"readerName":       "__" + structName + "Reader",
			"propType":         propType,
			"namesForType":     namesForType,
			"goLangType":       goLangType,
			"zeroValue":        zeroValue,
			"outputReadSetter": outputReadSetter,
		})
	}

	return output

}

func readerForStruct(useReflection bool, structName string) string {

	if useReflection {
		return templateOutput(reflectReaderTemplate, templateConfig{
			FirstLetter: structName[:1],
			StructName:  structName,
		})
	}

	return templateOutput(readerTemplate, map[string]string{
		"firstLetter": strings.ToLower(structName[:1]),
		"structName":  structName,
		"readerName":  "__" + structName + "Reader",
	})

}

func readSetterForStruct(useReflection bool, structName string) string {
	if useReflection {
		return templateOutput(reflectReadSetterTemplate, templateConfig{
			FirstLetter: structName[:1],
			StructName:  structName,
		})
	}

	return templateOutput(readSetterTemplate, map[string]string{
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

const reflectImportText = `import (
	"github.com/jkomoros/boardgame"
)

`

const importText = `import (
	"errors"
	"github.com/jkomoros/boardgame"
)

`

const reflectStructHeaderTemplateText = `// Implementation for {{.structName}}

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
	{{range $type, $goLangType := .propertyTypes -}}
	case boardgame.Type{{$type}}:
		val, ok := value.({{$goLangType}})
		if !ok {
			return errors.New("Provided value was not of type {{$goLangType}}")
		}
		return {{$firstLetter}}.Set{{$type}}Prop(name, val)
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

{{if .outputReadSetter -}}
func ({{.firstLetter}} *{{.readerName}}) Set{{.propType}}Prop(name string, value {{.goLangType}}) error {
	{{if .namesForType}}
	switch name {
		{{range .namesForType -}}
			case "{{.}}":
				{{$firstLetter}}.data.{{.}} = value
				return nil
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.propType}} prop: " + name)

}

{{end}}
`

const readerTemplateText = `func ({{.firstLetter}} *{{.structName}}) Reader() boardgame.PropertyReader {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const reflectReaderTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader({{.FirstLetter}})
}

`

const readSetterTemplateText = `func ({{.firstLetter}} *{{.structName}}) ReadSetter() boardgame.PropertyReadSetter {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const reflectReadSetterTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter({{.FirstLetter}})
}

`
