/*

	codegen is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter, as well as
	generating the boilerplate for enums.

	You typically don't use this package directly, but instead use the
	`boardgame-util codegen` command. See `boardgam-util help codegen` for
	more.

*/
package codegen

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
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

func process(options *appOptions, out io.ReadWriter, errOut io.ReadWriter) {

	//This is superceded by the codegen_cmd.go in boardgame-util, is here
	//mainly for testing purposes.

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
		ioutil.WriteFile(filepath.Join(options.PackageDirectory, options.OutputFile), []byte(output), 0644)
	}

	if testOutput != "" && options.OutputReaderTest {
		ioutil.WriteFile(filepath.Join(options.PackageDirectory, options.OutputFileTest), []byte(testOutput), 0644)
	}

	if enumOutput != "" && options.OutputEnum {
		ioutil.WriteFile(filepath.Join(options.PackageDirectory, options.EnumOutputFile), []byte(enumOutput), 0644)
	}

}

/*

ProcessPackage is a wrapper around ProcessStructs and ProcessEnums. It formats
the bytes before returning them.

*/
func processPackage(location string) (output string, testOutput string, enumOutput string, err error) {

	output, testOutput, err = ProcessStructs(location)

	if err != nil {
		return "", "", "", errors.New("Couldn't process structs: " + err.Error())
	}

	enumOutput, err = ProcessEnums(location)

	if err != nil {
		return "", "", "", errors.New("Couldn't process enums: " + err.Error())
	}

	return output, testOutput, enumOutput, nil

}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}
