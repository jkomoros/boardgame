package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/codegen"
	"io/ioutil"
	"path/filepath"
)

type Codegen struct {
	baseSubCommand

	PackageDirectory     string
	OutputFile           string
	OutputFileTest       string
	EnumOutputFile       string
	DontOutputEnum       bool
	DontOutputReader     bool
	DontOutputReaderTest bool
}

func (c *Codegen) Run(p writ.Path, positional []string) {
	output, testOutput, enumOutput, err := codegen.ProcessPackage(c.PackageDirectory)

	if err != nil {
		errAndQuit(err.Error())
	}

	if output != "" && !c.DontOutputReader {
		ioutil.WriteFile(filepath.Join(c.PackageDirectory, c.OutputFile), []byte(output), 0644)
	}

	if testOutput != "" && !c.DontOutputReaderTest {
		ioutil.WriteFile(filepath.Join(c.PackageDirectory, c.OutputFileTest), []byte(testOutput), 0644)
	}

	if enumOutput != "" && !c.DontOutputEnum {
		ioutil.WriteFile(filepath.Join(c.PackageDirectory, c.EnumOutputFile), []byte(enumOutput), 0644)
	}
}

func (c *Codegen) Name() string {
	return "codegen"
}

func (c *Codegen) Description() string {
	return "Automatically generates code to satisfy PropertyReader and generate Enum boilerplate"
}

func (c *Codegen) HelpText() string {
	return c.Name() +

		` automatically generates boilerplate PropertyReader and enums based
on structs in your package.

You can configure which package to process and where to write output via
command-line flags. By default it processes the current package and writes its
output to auto_reader.go, overwriting whatever file was there before. See
command-line options by passing -h. Structs with an boardgame:codegen comment that
are in a _test.go file will be outputin auto_reader_test.go.

See 'boardgame-util/lib/codegen' for more on its behavior.

The defaults are set reasonably so that you can use go:generate very
easily. See examplepkg/ for a very simple example.`

}

func (c *Codegen) WritOptions() []*writ.Option {

	return []*writ.Option{
		{
			Names: []string{"pkg"},
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&c.PackageDirectory),
				".",
			),
			Description: "Which package to process",
		},
		{
			Names: []string{"out"},
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&c.OutputFile),
				"auto_reader.go",
			),
			Description: "Defines which file to render output to. WARNING: it will be overwritten!",
		},
		{
			Names: []string{"outtest"},
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&c.OutputFileTest),
				"auto_reader_test.go",
			),
			Description: "For structs in files that end in _test.go, what is the filename they should be exported to?",
		},
		{
			Names: []string{"enumout"},
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&c.EnumOutputFile),
				"auto_enum.go",
			),
			Description: "Where to output the auto-enum file. WARNING: it will be overwritten!",
		},
		{
			Names:       []string{"no-enum"},
			Decoder:     writ.NewFlagDecoder(&c.DontOutputEnum),
			Description: "Whether to suppress output of auto_enum.go",
		},
		{
			Names:       []string{"no-reader"},
			Decoder:     writ.NewFlagDecoder(&c.DontOutputReader),
			Description: "Whether to suppress output of auto_reader.go",
		},
		{
			Names:       []string{"no-reader-test"},
			Decoder:     writ.NewFlagDecoder(&c.DontOutputReaderTest),
			Description: "Whether to suppress output of auto_reader_test.go",
		},
	}
}
