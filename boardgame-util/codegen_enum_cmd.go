package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/codegen"
	"io/ioutil"
	"path/filepath"
)

type CodegenEnum struct {
	baseSubCommand
}

func (c *CodegenEnum) Run(p writ.Path, positional []string) {

	packageDir := codegenPackageNameOrErr(positional)

	parent := c.Parent().(*Codegen)

	enumOutput, err := codegen.ProcessEnums(packageDir)

	if err != nil {
		errAndQuit("Couldn't process enums: " + err.Error())
	}

	if enumOutput != "" {
		if err := ioutil.WriteFile(filepath.Join(packageDir, parent.EnumOutputFile), []byte(enumOutput), 0644); err != nil {
			errAndQuit("Couldn't output file: " + err.Error())
		}
	}

}

func (c *CodegenEnum) Usage() string {
	return "PKGNAME"
}

func (c *CodegenEnum) Name() string {
	return "enum"
}

func (c *CodegenEnum) Description() string {
	return "Automatically generates enum boilerplate for a package"
}

func (c *CodegenEnum) HelpText() string {
	return c.Name() +

		` processes the given package and outputs the contents of a file
representing the auto-generated boilerplate for those enums. If it finds a
const() block at the top-level decorated with the magic comment (boardgame:codegen)
it will generate enum boilerplate. See the package doc of enum for more
on what you need to include.

auto-generated enums will automatically have values like PrefixVeryLongName
set to have a string value of "Very Long Name"; that is title-case will be
taken to mean word boundaries. If you want to transform the created values to
lowercase or uppercase, include a line of 'transform:lower' or
'transform:upper', respectively, in the comment lines immediately before the
constant. 'transform:none' means default behavior, leave as title case. If you
want to change the default transform for an entire const group, have the
transform line in the comment block above the constant block.  If you want to
override a specific item in the enum's name, include a comment immediately
above that matches that pattern 'display:"myVal"', where myVal is the exact
string to use. myVal may be zero-length, and may include quoted quotes. If
your enum has a key that is named with the prefix of the rest of the enum
values, and evaluates to 0, then a TreeEnum will be created. See the
documentation in the enum package for how to control nesting in a TreeEnum. If
enums are autogenerated, and the struct in your package that appears to be
your gameDelegate doesn't already have a ConfigureEnums(), one will be
generated for you.

If PKGNAME parameter is missing, "." is assumed.

This command is a thin wrapper around 'boardgame-util/lib/codegen.ProcessEnums'.
`

}