package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
	codegenPkg "github.com/jkomoros/boardgame/boardgame-util/lib/codegen"
)

type codegenReader struct {
	baseSubCommand

	DontOutputReaderTest bool
}

func (c *codegenReader) Run(p writ.Path, positional []string) {

	pkgDirectory := dirPositionalOrDefault(c.Base(), positional, true)

	parent := c.Parent().(*codegen)

	readerOutput, testReaderOutput, err := codegenPkg.ProcessReaders(pkgDirectory)

	if err != nil {
		c.Base().errAndQuit("Couldn't process readers: " + err.Error())
	}

	outputs := make(map[string][]byte, 2)
	if readerOutput != "" {
		outputs[parent.OutputFile] = []byte(readerOutput)
	}
	if !c.DontOutputReaderTest && testReaderOutput != "" {
		outputs[parent.OutputFileTest] = []byte(testReaderOutput)
	}
	if err := fileutil.WriteFilesAtomic(pkgDirectory, outputs, true, 0o644); err != nil {
		c.Base().errAndQuit("Couldn't publish generated reader files: " + err.Error())
	}

}

func (c *codegenReader) Name() string {
	return "reader"
}

func (c *codegenReader) Description() string {
	return "Automatically generates PropertyReader boilerplate for a package"
}

func (c *codegenReader) Usage() string {
	return "PKGNAME"
}

func (c *codegenReader) HelpText() string {
	return c.Name() +

		` generates PropertyReader and friends for a given package.

reader processes a package of go files, searching for structs that
have a comment immediately above their declaration that begins with
"boardgame:codegen". For each such struct, it creates a Reader(), ReadSetter(),
and ReadSetConfigurer() method that implement boardgame.Reader,
boardgame.ReadSetter, and boardgame.ReadSetConfigurer, respectively.

Producing a ReadSetConfigurator requires a ReadSetter, and producing a
ReadSetter requires a Reader. By default if you have the magic comment of
'boardgame:codegen' it with produce all three. However, if you want only some of
the methods, include an argument for the highest one you want, e.g.
'boardgame:codegen readsetter' to generate a Reader() and ReadSetter().

This package will automatically create additional type transform methods
to handle fields whose literal type is boardgame.ImmutableSizedStack,
boardgame.SizedStack, boardgame.MergedStack, enum.RangeValue, and
enum.TreeValue.

Structs with an boardgame:codegen comment that
are in a _test.go file will be outputin auto_reader_test.go.

The outputted readers, readsetters, and readsetconfigurers use a hard-
coded list of fields for performance (reflection would be about 30% slower
under normal usage). You should re-generate output every time you add a
struct or modify the fields on a struct.

If PKGNAME parameter is missing, "." is assumed.

It is a thin wrapper around 'boardgame- util/lib/codegen.ProcessReaders'. `

}

func (c *codegenReader) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"no-test"},
			Decoder:     writ.NewFlagDecoder(&c.DontOutputReaderTest),
			Flag:        true,
			Description: "If provided, won't output auto_reader_test.go even if content exists for it",
		},
	}
}
