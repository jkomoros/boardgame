package main

import (
	"github.com/bobziuchkovski/writ"
)

type Codegen struct {
	baseSubCommand

	CodegenAll    CodegenAll
	CodegenEnum   CodegenEnum
	CodegenReader CodegenReader

	OutputFile     string
	OutputFileTest string
	EnumOutputFile string
}

func (c *Codegen) Run(p writ.Path, positional []string) {
	c.CodegenAll.Run(p, positional)
}

func (c *Codegen) Name() string {
	return "codegen"
}

func codegenPackageNameOrErr(positional []string) string {
	if len(positional) > 1 {
		errAndQuit("More than one positional argument provided, expecting only package")
	}
	if len(positional) == 0 {
		return "."
	}
	return positional[0]
}

func (c *Codegen) Description() string {
	return "Automatically generates code to satisfy PropertyReader and generate Enum boilerplate"
}

func (c *Codegen) HelpText() string {
	return c.Name() +

		` automatically generates boilerplate PropertyReader and enums based
on structs in your package.

Running this command and not any of its subcommands is equivalent to running
'codegen all'. If PKGNAME parameter is missing, "." is assumed.

See 'boardgame-util/lib/codegen' for more on its behavior.

The defaults are set reasonably so that you can use it easily with go generate
by including the following line in your package:

` + "//go:" + "generate boardgame-util codegen" + `

See examplepkg/ for a very simple example.`

}

func (c *Codegen) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.CodegenAll,
		&c.CodegenReader,
		&c.CodegenEnum,
	}
}

func (c *Codegen) Usage() string {
	return "PKGNAME"
}

func (c *Codegen) WritOptions() []*writ.Option {

	return []*writ.Option{
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
	}
}
