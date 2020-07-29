package codegen

import (
	"strings"
	"text/template"
)

var headerTemplate *template.Template
var structHeaderTemplate *template.Template
var typedPropertyTemplate *template.Template
var readerTemplate *template.Template
var readSetterTemplate *template.Template
var readSetConfigurerTemplate *template.Template
var enumHeaderTemplate *template.Template
var enumDelegateTemplate *template.Template
var enumItemTemplate *template.Template
var enumGroupTemplate *template.Template

func init() {
	funcMap := template.FuncMap{
		"immutablekey": immutableKey,
		"verbfortype":  verbForType,
		"firstLetter":  firstLetter,
	}

	headerTemplate = template.Must(template.New("header").Funcs(funcMap).Parse(headerTemplateText))
	structHeaderTemplate = template.Must(template.New("structHeader").Funcs(funcMap).Parse(structHeaderTemplateText))
	typedPropertyTemplate = template.Must(template.New("typedProperty").Funcs(funcMap).Parse(typedPropertyTemplateText))
	readerTemplate = template.Must(template.New("reader").Funcs(funcMap).Parse(readerTemplateText))
	readSetterTemplate = template.Must(template.New("readsetter").Funcs(funcMap).Parse(readSetterTemplateText))
	readSetConfigurerTemplate = template.Must(template.New("readsetconfigurer").Funcs(funcMap).Parse(readSetConfigurerTemplateText))
	enumHeaderTemplate = template.Must(template.New("enumheader").Funcs(funcMap).Parse(enumHeaderTemplateText))
	enumDelegateTemplate = template.Must(template.New("enumdelegate").Funcs(funcMap).Parse(enumDelegateTemplateText))
	enumItemTemplate = template.Must(template.New("enumitem").Parse(enumItemTemplateText))
	enumGroupTemplate = template.Must(template.New("enumgroup").Parse(enumGroupTemplateText))
}

func immutableKey(in propertyType) string {
	var result string
	if in.IsInterface() {
		result = "Immutable"
	}
	result += in.Key()
	return result
}

func verbForType(in propertyType) string {
	if in.IsInterface() {
		return "Configure"
	}
	return "Set"
}

func firstLetter(in string) string {

	if in == "" {
		return ""
	}

	return strings.ToLower(in[:1])
}

const headerTemplateText = `/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated 
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package {{.packageName}}
`

const importText = `import (
	"errors"
	
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

`

const structHeaderTemplateText = `// Implementation for {{.StructName}}

var {{.ReaderName}}Props = map[string]boardgame.PropertyType{
	{{range $key, $value := .Fields -}}
		"{{$key}}": boardgame.{{$value.Type.String}},
	{{end}}
}

type {{.ReaderName}} struct {
	data *{{.StructName}}
}

func ({{.FirstLetter}} *{{.ReaderName}}) Props() map[string]boardgame.PropertyType {
	return {{.ReaderName}}Props
}

func ({{.FirstLetter}} *{{.ReaderName}}) Prop(name string) (interface{}, error) {
	props := {{.FirstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	{{$firstLetter := .FirstLetter}}

	switch propType {
	{{range $type := .PropertyTypes -}}
	case boardgame.Type{{$type.Key}}:
		return {{$firstLetter}}.{{immutablekey $type}}Prop(name)
	{{end}}
	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

{{if .OutputReadSetter -}}

func ({{.FirstLetter}} *{{.ReaderName}}) PropMutable(name string) bool {
	switch name {
		{{range $key, $val := .Fields -}}
	case "{{$key}}":
		return {{$val.Mutable}}
		{{end -}}
	}

	return false
}

func ({{.FirstLetter}} *{{.ReaderName}}) SetProp(name string, value interface{}) error {
	props := {{.FirstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	{{range $type := .PropertyTypes -}}
	{{if $type.IsInterface -}}
	case boardgame.Type{{$type.Key}}:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	{{- else -}}
	case boardgame.Type{{$type.Key}}:
		val, ok := value.({{$type.ImmutableGoType}})
		if !ok {
			return errors.New("Provided value was not of type {{$type.ImmutableGoType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{$type.Key}}Prop(name, val)
	{{- end}}
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}

{{if .OutputReadSetConfigurer -}}
func ({{.FirstLetter}} *{{.ReaderName}}) ConfigureProp(name string, value interface{}) error {
	props := {{.FirstLetter}}.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	{{range $type := .PropertyTypes -}}
	case boardgame.Type{{$type.Key}}:
		{{if $type.IsInterface -}}
		if {{$firstLetter}}.PropMutable(name) {
			//Mutable variant
			val, ok := value.({{$type.MutableGoType}})
			if !ok {
				return errors.New("Provided value was not of type {{$type.MutableGoType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type.Key}}Prop(name, val)
		}
		//Immutable variant
		val, ok := value.({{$type.ImmutableGoType}})
		if !ok {
			return errors.New("Provided value was not of type {{$type.ImmutableGoType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{immutablekey $type}}Prop(name, val)
		{{- else -}}
			val, ok := value.({{$type.ImmutableGoType}})
			if !ok {
				return errors.New("Provided value was not of type {{$type.ImmutableGoType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type.Key}}Prop(name, val)
		{{- end}}
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}
`

const typedPropertyTemplateText = `func ({{.FirstLetter}} *{{.ReaderName}}) {{immutablekey .PropType}}Prop(name string) ({{.PropType.ImmutableGoType}}, error) {
	{{$firstLetter := .FirstLetter}}
	{{if .FieldsForType}}
	switch name {
		{{range .FieldsForType -}}
			case "{{.Name}}":
				return {{$firstLetter}}.data.{{.Name}}, nil
		{{end}}
	}
	{{end}}

	return {{.PropType.ZeroValue}}, errors.New("No such {{.PropType.Key}} prop: " + name)

}

{{if .OutputReadSetConfigurer -}}
{{if .PropType.IsInterface -}}
func ({{.FirstLetter}} *{{.ReaderName}}) Configure{{.PropType.Key}}Prop(name string, value {{.PropType.MutableGoType}}) error {
	{{if .FieldsForType}}
	switch name {
		{{range .FieldsForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				{{if .SubType -}}
				slotValue := value.{{.SubType}}()
				if slotValue == nil {
					return errors.New("{{.Name}} couldn't be upconverted, returned nil")
				}
				{{$firstLetter}}.data.{{.Name}} = slotValue
				{{- else -}}
				{{$firstLetter}}.data.{{.Name}} = value
				{{- end}}
				return nil
			{{- else -}}
				return boardgame.ErrPropertyImmutable
			{{- end}}
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.PropType.Key}} prop: " + name)

}

func ({{.FirstLetter}} *{{.ReaderName}}) Configure{{immutablekey .PropType}}Prop(name string, value {{.PropType.ImmutableGoType}}) error {
	{{if .FieldsForType}}
	switch name {
		{{range .FieldsForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return boardgame.ErrPropertyImmutable
			{{- else -}}
				{{if .SubType -}}
				slotValue := value.{{.SubType}}()
				if slotValue == nil {
					return errors.New("{{.Name}} couldn't be upconverted, returned nil")
				}
				{{$firstLetter}}.data.{{.Name}} = slotValue
				{{- else -}}
				{{$firstLetter}}.data.{{.Name}} = value
				{{- end}}
				return nil
			{{- end}}
		{{end}}
	}
	{{end}}

	return errors.New("No such {{immutablekey .PropType}} prop: " + name)

}

{{end}}
{{end}}

{{if .OutputReadSetter -}}
{{if .PropType.IsInterface -}}
func ({{.FirstLetter}} *{{.ReaderName}}) {{.PropType.Key}}Prop(name string) ({{.PropType.MutableGoType}}, error) {
	{{$firstLetter := .FirstLetter}}
	{{$zeroValue := .PropType.ZeroValue}}
	{{if .FieldsForType}}
	switch name {
		{{range .FieldsForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return {{$firstLetter}}.data.{{.Name}}, nil
			{{- else -}}
				return {{$zeroValue}}, boardgame.ErrPropertyImmutable
			{{- end}}
		{{end}}
	}
	{{end}}

	return {{.PropType.ZeroValue}}, errors.New("No such {{.PropType.Key}} prop: " + name)

}

{{else}}
func ({{.FirstLetter}} *{{.ReaderName}}) Set{{.PropType.Key}}Prop(name string, value {{.PropType.MutableGoType}}) error {
	{{if .FieldsForType}}
	switch name {
		{{range .FieldsForType -}}
			case "{{.Name}}":
				{{$firstLetter}}.data.{{.Name}} = value
				return nil
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.PropType.Key}} prop: " + name)

}

{{end}}
{{end}}
`

const readerTemplateText = `//Reader returns an autp-generated boardgame.PropertyReader for {{.StructName}}
func ({{.FirstLetter}} *{{.StructName}}) Reader() boardgame.PropertyReader {
	return &{{.ReaderName}}{ {{.FirstLetter}} }
}

`

const readSetterTemplateText = `//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for {{.StructName}}
func ({{.FirstLetter}} *{{.StructName}}) ReadSetter() boardgame.PropertyReadSetter {
	return &{{.ReaderName}}{ {{.FirstLetter}} }
}

`

const readSetConfigurerTemplateText = `//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for {{.StructName}}
func ({{.FirstLetter}} *{{.StructName}}) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &{{.ReaderName}}{ {{.FirstLetter}} }
}

`
const enumHeaderTemplateText = `/************************************
 *
 * This file contains auto-generated methods to help configure enums. 
 * It was generated by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package {{.packageName}}

import (
	"github.com/jkomoros/boardgame/enum"
)

var enums = enum.NewSet()

`

const enumDelegateTemplateText = `//ConfigureEnums simply returns enums, the auto-generated Enums variable. This
//is output because {{.delegateName}} appears to be a struct that implements
//boardgame.GameDelegate, and does not already have a ConfigureEnums
//explicitly defined.
func ({{firstLetter .delegateName}} *{{.delegateName}}) ConfigureEnums() *enum.Set {
	return enums
}

`

const enumItemTemplateText = `{{if .firstNewKey}} 
//Implicitly created constants for {{.prefix}}
const (
	{{.firstNewKey}} = iota - 9223372036854775808
{{range .restNewKeys -}}
	{{.}}
{{- end -}}
)

{{ end -}}
//{{.prefix}}Enum is the enum.Enum for {{.prefix}}
var {{.prefix}}Enum = enums.MustAdd{{if .parents}}Tree{{end}}("{{.prefix}}", map[int]string{
	{{ $prefix := .prefix -}}
	{{range $name, $value := .values -}}
	{{$name}}: "{{$value}}",
	{{end}}
{{if .parents -}} }, map[int]int{ 
	{{ $prefix := .prefix -}}
	{{range $name, $value := .parents -}}
	{{$name}}: {{$value}},
	{{end}}
{{end -}}
})

`

const enumGroupTemplateText = `//{{.name}}Enum is the combined enum.Enum for {{.name}}
var {{.name}}Enum = enums.MustCombine("{{.name}}", {{.varNames}})

`
