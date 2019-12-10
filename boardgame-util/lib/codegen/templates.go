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

func init() {
	funcMap := template.FuncMap{
		"withimmutable": withImmutable,
		"ismutable":     isMutable,
		"verbfortype":   verbForType,
		"firstLetter":   firstLetter,
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
}

func withImmutable(in string) string {
	prefix := ""
	rest := in
	parts := strings.Split(in, ".")
	if len(parts) > 1 {
		prefix = strings.Join(parts[:len(parts)-1], ".")
		rest = parts[len(parts)-1]
	}

	if _, needsImmutable := configureTypes[rest]; !needsImmutable {
		return in
	}

	rest = "Immutable" + rest

	if prefix == "" {
		return rest
	}

	return prefix + "." + rest

}

func isMutable(in string) bool {

	for key := range configureTypes {
		if strings.Contains(in, key) {
			return true
		}
	}

	return false
}

var configureTypes = map[string]bool{
	"Stack": true,
	"Timer": true,
	"Board": true,
	"Val":   true,
	"Enum":  true,
}

func verbForType(in string) string {
	_, configure := configureTypes[in]
	if configure {
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
	{{range $key, $value := .Fields.Types -}}
		"{{$key}}": boardgame.{{$value.String}},
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
	{{range $type, $goLangtype := .PropertyTypes -}}
	case boardgame.Type{{$type}}:
		return {{$firstLetter}}.{{withimmutable $type}}Prop(name)
	{{end}}
	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

{{if .OutputReadSetter -}}

func ({{.FirstLetter}} *{{.ReaderName}}) PropMutable(name string) bool {
	switch name {
		{{range $key, $val := .Fields.Mutable -}}
	case "{{$key}}":
		return {{$val}}
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
	{{range $type, $goLangType := .SetterPropertyTypes -}}
	{{if ismutable $type -}}
	case boardgame.Type{{$type}}:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	{{- else -}}
	case boardgame.Type{{$type}}:
		val, ok := value.({{$goLangType}})
		if !ok {
			return errors.New("Provided value was not of type {{$goLangType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
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
	{{range $type, $goLangType := .SetterPropertyTypes -}}
	case boardgame.Type{{$type}}:
		{{if ismutable $type -}}
		if {{$firstLetter}}.PropMutable(name) {
			//Mutable variant
			val, ok := value.({{$goLangType}})
			if !ok {
				return errors.New("Provided value was not of type {{$goLangType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
		}
		//Immutable variant
		val, ok := value.({{withimmutable $goLangType}})
		if !ok {
			return errors.New("Provided value was not of type {{withimmutable $goLangType}}")
		}
		return {{$firstLetter}}.{{verbfortype $type}}{{withimmutable $type}}Prop(name, val)
		{{- else -}}
			val, ok := value.({{$goLangType}})
			if !ok {
				return errors.New("Provided value was not of type {{$goLangType}}")
			}
			return {{$firstLetter}}.{{verbfortype $type}}{{$type}}Prop(name, val)
		{{- end}}
	{{end}}
	}

	return errors.New("Unexpected property type: " + propType.String())
}

{{end}}
`

const typedPropertyTemplateText = `func ({{.FirstLetter}} *{{.ReaderName}}) {{withimmutable .PropType}}Prop(name string) ({{.GoLangType}}, error) {
	{{$firstLetter := .FirstLetter}}
	{{if .NamesForType}}
	switch name {
		{{range .NamesForType -}}
			case "{{.Name}}":
				return {{$firstLetter}}.data.{{.Name}}, nil
		{{end}}
	}
	{{end}}

	return {{.ZeroValue}}, errors.New("No such {{.PropType}} prop: " + name)

}

{{if .OutputReadSetConfigurer -}}
{{if .OutputMutableGetter -}}
func ({{.FirstLetter}} *{{.ReaderName}}) Configure{{.SetterPropType}}Prop(name string, value {{.SetterGoLangType}}) error {
	{{if .NamesForType}}
	switch name {
		{{range .NamesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				{{if .UpConverter -}}
				slotValue := value.{{.UpConverter}}()
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

	return errors.New("No such {{.SetterPropType}} prop: " + name)

}

func ({{.FirstLetter}} *{{.ReaderName}}) Configure{{withimmutable .SetterPropType}}Prop(name string, value {{withimmutable .SetterGoLangType}}) error {
	{{if .NamesForType}}
	switch name {
		{{range .NamesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return boardgame.ErrPropertyImmutable
			{{- else -}}
				{{if .UpConverter -}}
				slotValue := value.{{.UpConverter}}()
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

	return errors.New("No such {{withimmutable .SetterPropType}} prop: " + name)

}

{{end}}
{{end}}

{{if .OutputReadSetter -}}
{{if .OutputMutableGetter -}}
func ({{.FirstLetter}} *{{.ReaderName}}) {{.SetterPropType}}Prop(name string) ({{.SetterGoLangType}}, error) {
	{{$firstLetter := .FirstLetter}}
	{{$zeroValue := .ZeroValue}}
	{{if .NamesForType}}
	switch name {
		{{range .NamesForType -}}
			case "{{.Name}}":
			{{if .Mutable -}}
				return {{$firstLetter}}.data.{{.Name}}, nil
			{{- else -}}
				return {{$zeroValue}}, boardgame.ErrPropertyImmutable
			{{- end}}
		{{end}}
	}
	{{end}}

	return {{.ZeroValue}}, errors.New("No such {{.PropType}} prop: " + name)

}

{{else}}
func ({{.FirstLetter}} *{{.ReaderName}}) Set{{.SetterPropType}}Prop(name string, value {{.SetterGoLangType}}) error {
	{{if .NamesForType}}
	switch name {
		{{range .NamesForType -}}
			case "{{.Name}}":
				{{$firstLetter}}.data.{{.Name}} = value
				return nil
		{{end}}
	}
	{{end}}

	return errors.New("No such {{.SetterPropType}} prop: " + name)

}

{{end}}
{{end}}
`

const readerTemplateText = `//Reader returns an autp-generated boardgame.PropertyReader for {{.structName}}
func ({{.firstLetter}} *{{.structName}}) Reader() boardgame.PropertyReader {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const readSetterTemplateText = `//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for {{.structName}}
func ({{.firstLetter}} *{{.structName}}) ReadSetter() boardgame.PropertyReadSetter {
	return &{{.readerName}}{ {{.firstLetter}} }
}

`

const readSetConfigurerTemplateText = `//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for {{.structName}}
func ({{.firstLetter}} *{{.structName}}) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &{{.readerName}}{ {{.firstLetter}} }
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
