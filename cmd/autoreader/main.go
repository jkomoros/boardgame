/*

	Autoreader is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter.

*/
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/MarcGrol/golangAnnotations/parser"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

var headerTemplate *template.Template
var readerTemplate *template.Template
var readSetterTemplate *template.Template

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
	headerTemplate = template.Must(template.New("header").Parse(headerTemplateText))
	readerTemplate = template.Must(template.New("reader").Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Parse(readSetterTemplateText))
}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.PackageDirectory, "pkg", "examplepkg/", "Which package to process")
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

		enableAutoReader := false

		for _, docLine := range theStruct.DocLines {
			docLine = strings.TrimPrefix(docLine, "//")
			docLine = strings.TrimSpace(docLine)
			if strings.HasPrefix(docLine, magicDocLinePrefix) {
				enableAutoReader = true
				break
			}
		}

		if !enableAutoReader {
			continue
		}

		output += readerForStruct(theStruct.Name)
		output += readSetterForStruct(theStruct.Name)
	}

	return output, nil
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
 * This file was auto-generated by autoreader. DO NOT EDIT.
 *
 ************************************/
package {{.packageName}}

import (
	"github.com/jkomoros/boardgame"
)

`

const readerTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader({{.FirstLetter}})
}

`

const readSetterTemplateText = `func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter({{.FirstLetter}})
}

`
