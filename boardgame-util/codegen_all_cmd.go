package main

import (
	"github.com/bobziuchkovski/writ"
)

type CodegenAll struct {
	baseSubCommand
}

func (c *CodegenAll) Run(p writ.Path, positional []string) {

	parent := c.Parent().(*Codegen)

	parent.CodegenReader.Run(p, positional)
	parent.CodegenEnum.Run(p, positional)

}

func (c *CodegenAll) Name() string {
	return "all"
}

func (c *CodegenAll) Description() string {
	return "Automatically generates PropertyReader and enum boilerplate for a package"
}

func (c *CodegenAll) HelpText() string {
	return c.Name() +

		` generates both PropertyReader and enum output for the given package.
It is equivalent to 'codegen reader' followed by 'codegen enum'. The
'codegen' command without any sub-commands defaults to 'codegen all'.

If PKGNAME parameter is missing, "." is assumed.

For more on the command, see the help for 'codegen reader' and 'codegen enum'.
`

}
