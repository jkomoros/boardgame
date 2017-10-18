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
	"github.com/MarcGrol/golangAnnotations/parser"
	"github.com/jkomoros/boardgame"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type memoizedEmbeddedStructKey struct {
	Import           string
	TargetStructName string
}

var memoizedEmbeddedStructs map[memoizedEmbeddedStructKey]map[string]boardgame.PropertyType

const magicDocLinePrefix = "+autoreader"

type appOptions struct {
	OutputFile       string
	EnumOutputFile   string
	PackageDirectory string
	PrintToConsole   bool
	Help             bool
	flagSet          *flag.FlagSet
}

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.EnumOutputFile, "enumout", "auto_enum.go", "Where to output the auto-enum file. WARNING: it will be overwritten!")
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

	output, enumOutput, err := processPackage(options.PackageDirectory)

	if err != nil {
		fmt.Fprintln(errOut, "ERROR", err)
		return
	}

	if options.PrintToConsole {
		fmt.Fprintln(out, output)
		fmt.Fprintln(out, enumOutput)
		return
	}

	if output != "" {
		ioutil.WriteFile(options.OutputFile, []byte(output), 0644)
	}

	if enumOutput != "" {
		ioutil.WriteFile(options.EnumOutputFile, []byte(enumOutput), 0644)
	}

}

func processPackage(location string) (output string, enumOutput string, err error) {

	sources, err := parser.ParseSourceDir(location, ".*")

	if err != nil {
		return "", "", errors.New("Couldn't parse sources: " + err.Error())
	}

	output, err = processStructs(sources, location)

	if err != nil {
		return "", "", errors.New("Couldn't process structs: " + err.Error())
	}

	enumOutput, err = processEnums(sources)

	if err != nil {
		return "", "", errors.New("Couldn't process enums: " + err.Error())
	}

	formattedBytes, err := format.Source([]byte(output))

	if err != nil {
		return "", "", errors.New("Couldn't go fmt code for reader: " + err.Error())
	}

	formattedEnumBytes, err := format.Source([]byte(enumOutput))

	if err != nil {
		return "", "", errors.New("Couldn't go fmt code for enums: " + err.Error())
	}

	return string(formattedBytes), string(formattedEnumBytes), nil

}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}
