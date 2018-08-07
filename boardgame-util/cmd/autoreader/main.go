/*

	Autoreader is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter, as well as
	generating the boilerplate for enums.

	You can configure which package to process and where to write output via
	command-line flags. By default it processes the current package and writes
	its output to auto_reader.go, overwriting whatever file was there before.
	See command-line options by passing -h. Structs with an +autoreader
	comment that are in a _test.go file will be outputin auto_reader_test.go.

	The defaults are set reasonably so that you can use go:generate very
	easily. See examplepkg/ for a very simple example.

	You typically don't use this package directly, but instead use the
	`boardgame-util codegen` command. See `boardgam-util help codegen` for
	more.

*/
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

//debugSaveBadCode, if true, will save even code that is not legal go if it
//can't be formatted. Useful for debugging bad template output temporarily.
//Should never be set to true for real uses.
const debugSaveBadCode = false

type appOptions struct {
	OutputFile       string
	OutputFileTest   string
	EnumOutputFile   string
	PackageDirectory string
	PrintToConsole   bool
	OutputEnum       bool
	OutputReader     bool
	OutputReaderTest bool
	Help             bool
	flagSet          *flag.FlagSet
}

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.OutputFile, "out", "auto_reader.go", "Defines which file to render output to. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.OutputFileTest, "outtest", "auto_reader_test.go", "For structs in files that end in _test.go, what is the filename they should be exported to?")
	options.flagSet.StringVar(&options.EnumOutputFile, "enumout", "auto_enum.go", "Where to output the auto-enum file. WARNING: it will be overwritten!")
	options.flagSet.StringVar(&options.PackageDirectory, "pkg", ".", "Which package to process")
	options.flagSet.BoolVar(&options.OutputEnum, "enum", true, "Whether or not to output auto_enum.go")
	options.flagSet.BoolVar(&options.OutputReader, "reader", true, "Whether or not to output auto_reader.go")
	options.flagSet.BoolVar(&options.OutputReaderTest, "readertest", true, "Whether or not to output auto_reader_test.go")
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

	output, testOutput, enumOutput, err := processPackage(options.PackageDirectory)

	if err != nil {
		fmt.Fprintln(errOut, "ERROR", err)
		return
	}

	if options.PrintToConsole {
		if options.OutputReader {
			fmt.Fprintln(out, output)

		}
		if options.OutputReaderTest {
			fmt.Fprintln(out, testOutput)
		}
		if options.OutputEnum {
			fmt.Fprintln(out, enumOutput)
		}

		return
	}

	if output != "" && options.OutputReader {
		ioutil.WriteFile(options.OutputFile, []byte(output), 0644)
	}

	if testOutput != "" && options.OutputReaderTest {
		ioutil.WriteFile(options.OutputFileTest, []byte(testOutput), 0644)
	}

	if enumOutput != "" && options.OutputEnum {
		ioutil.WriteFile(options.EnumOutputFile, []byte(enumOutput), 0644)
	}

}

func processPackage(location string) (output string, testOutput string, enumOutput string, err error) {

	output, testOutput, err = ProcessStructs(location)

	if err != nil {
		return "", "", "", errors.New("Couldn't process structs: " + err.Error())
	}

	enumOutput, err = ProcessEnums(location)

	if err != nil {
		return "", "", "", errors.New("Couldn't process enums: " + err.Error())
	}

	formattedBytes, err := format.Source([]byte(output))

	if err != nil {
		if debugSaveBadCode {
			formattedBytes = []byte(output)
		} else {
			return "", "", "", errors.New("Couldn't go fmt code for reader: " + err.Error())
		}
	}

	formattedTestBytes, err := format.Source([]byte(testOutput))

	if err != nil {
		if debugSaveBadCode {
			formattedTestBytes = []byte(testOutput)
		} else {
			return "", "", "", errors.New("Couldn't go fmt code for reader: " + err.Error())
		}
	}

	formattedEnumBytes, err := format.Source([]byte(enumOutput))

	if err != nil {
		if debugSaveBadCode {
			formattedEnumBytes = []byte(enumOutput)
		} else {
			return "", "", "", errors.New("Couldn't go fmt code for enums: " + err.Error())
		}
	}

	return string(formattedBytes), string(formattedTestBytes), string(formattedEnumBytes), nil

}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}
