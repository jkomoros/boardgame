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
	"github.com/MarcGrol/golangAnnotations/parser"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

var headerTemplate *template.Template
var structHeaderTemplate *template.Template
var readerTemplate *template.Template
var readSetterTemplate *template.Template

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
	readerTemplate = template.Must(template.New("reader").Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Parse(readSetterTemplateText))
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
			output += headerForPackage(theStruct.PackageName)
			haveOutputHeader = true
		}

		outputReader, outputReadSetter := structConfig(theStruct.DocLines)

		if outputReader || outputReadSetter {
			output += headerForStruct(theStruct.Name)
		}

		if outputReader {
			output += readerForStruct(theStruct.Name)
		}
		if outputReadSetter {
			output += readSetterForStruct(theStruct.Name)
		}
	}

	formattedBytes, err := format.Source([]byte(output))

	if err != nil {
		return "", errors.New("Couldn't go fmt code: " + err.Error())
	}

	return string(formattedBytes), nil
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

func headerForPackage(packageName string) string {
	return templateOutput(headerTemplate, map[string]string{
		"packageName": packageName,
	}) + reflectImportText
}

func headerForStruct(structName string) string {
	return templateOutput(structHeaderTemplate, map[string]string{
		"structName": structName,
	})
}

func readerForStruct(structName string) string {

	return templateOutput(readerTemplate, templateConfig{
		FirstLetter: structName[:1],
		StructName:  structName,
	})

}

func readSetterForStruct(structName string) string {
	return templateOutput(readSetterTemplate, templateConfig{
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

const prodImportText = `import (
	"errors"
	"github.com/jkomoros/boardgame"
)

`

const structHeaderTemplateText = `// Implementation for {{.structName}}

 `

const readerTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader({{.FirstLetter}})
}

`

const readSetterTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter({{.FirstLetter}})
}

`
