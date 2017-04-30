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
	"strings"
	"text/template"
	"unicode"
)

var headerTemplate *template.Template
var structHeaderTemplate *template.Template
var reflectStructHeaderTemplate *template.Template
var reflectReaderTemplate *template.Template
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
	reflectStructHeaderTemplate = template.Must(template.New("structHeader").Parse(reflectStructHeaderTemplateText))
	reflectReaderTemplate = template.Must(template.New("reader").Parse(reflectReaderTemplateText))
	reflectReadSetterTemplate = template.Must(template.New("readsetter").Parse(reflectReadSetterTemplateText))

}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.PackageDirectory, "pkg", ".", "Which package to process")
	options.flagSet.BoolVar(&options.Help, "h", false, "If set, print help message and quit.")
	options.flagSet.BoolVar(&options.PrintToConsole, "print", false, "If true, will print result to console instead of writing to out.")
	options.flagSet.BoolVar(&options.UseReflection, "reflect", true, "If true, will use reflection based output.")
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
			output += headerForStruct(useReflection, theStruct.Name, types)
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

func headerForStruct(useReflection bool, structName string, types map[string]boardgame.PropertyType) string {

	if useReflection {
		return templateOutput(reflectStructHeaderTemplate, map[string]string{
			"structName": structName,
		})
	}

	return templateOutput(structHeaderTemplate, map[string]interface{}{
		"structName": structName,
		"types":      types,
	})

}

func readerForStruct(useReflection bool, structName string) string {

	return templateOutput(reflectReaderTemplate, templateConfig{
		FirstLetter: structName[:1],
		StructName:  structName,
	})

}

func readSetterForStruct(useReflection bool, structName string) string {
	return templateOutput(reflectReadSetterTemplate, templateConfig{
		FirstLetter: structName[:1],
		StructName:  structName,
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

var {{.structName}}ReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	{{range $key, $value := .types -}}
		"{{$key}}": boardgame.{{$value.String}},
	{{end}}
}

`

const reflectReaderTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader({{.FirstLetter}})
}

`

const reflectReadSetterTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter({{.FirstLetter}})
}

`
