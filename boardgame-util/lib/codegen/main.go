/*

Package codegen is a simple program, designed to be run from go:generate, that
helps generate the annoying boilerplate to implement boardgame.PropertyReader
and boardgame.PropertyReadSetter, as well as generating the boilerplate for
enums.

You typically don't use this package directly, but instead use the
`boardgame-util codegen` command. See `boardgam-util help codegen` for more.

*/
package codegen

import (
	"bytes"
	"log"
	"text/template"
)

//debugSaveBadCode, if true, will save even code that is not legal go if it
//can't be formatted. Useful for debugging bad template output temporarily.
//Should never be set to true for real uses.
const debugSaveBadCode = false

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func templateOutput(template *template.Template, values interface{}) string {
	buf := new(bytes.Buffer)

	err := template.Execute(buf, values)

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}
