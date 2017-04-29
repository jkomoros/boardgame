/*

	Autoreader is a simple program, designed to be run from go:generate, that
	helps generate the annoying boilerplate to implement
	boardgame.PropertyReader and boardgame.PropertyReadSetter.

*/
package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
)

var readerTemplate *template.Template
var readSetterTemplate *template.Template

type templateConfig struct {
	FirstLetter string
	StructName  string
}

func init() {
	readerTemplate = template.Must(template.New("reader").Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Parse(readSetterTemplateText))
}

func main() {
	fmt.Println(readerForStruct("myStruct"))
	fmt.Println(readSetterForStruct("myStruct"))
}

func readerForStruct(structName string) string {
	buf := new(bytes.Buffer)

	err := readerTemplate.Execute(buf, templateConfig{
		FirstLetter: structName[:1],
		StructName:  structName,
	})

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}

func readSetterForStruct(structName string) string {
	buf := new(bytes.Buffer)

	err := readSetterTemplate.Execute(buf, templateConfig{
		FirstLetter: structName[:1],
		StructName:  structName,
	})

	if err != nil {
		log.Println(err)
	}

	return buf.String()
}

const readerTemplateText = `
func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader({{.FirstLetter}})
}

`

const readSetterTemplateText = `
func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter({{.FirstLetter}})
}

`
