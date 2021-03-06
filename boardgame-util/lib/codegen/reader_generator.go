package codegen

import (
	"sort"
	"strings"

	"github.com/MarcGrol/golangAnnotations/model"
)

//readerGenerator represents a strucxt in the imported code that had the magic
//codegen tag attached, meaning that we should generate code for it.
type readerGenerator struct {
	s                       model.Struct
	outputReader            bool
	outputReadSetter        bool
	outputReadSetConfigurer bool
	//TODO: pop all of this directly into the struct
	fields fieldsInfo
}

//newReaderGenerator processes the given struct and then outputs a generator if
//any code is necessary to be output.
func newReaderGenerator(s model.Struct, location string, allStructs []model.Struct) *readerGenerator {
	outputReader, outputReadSetter, outputReadSetConfigurer := structConfig(s.DocLines)

	if !outputReader && !outputReadSetter && !outputReadSetConfigurer {
		return nil
	}

	//readSetConfigurer is a superset of readSetter, which means that if
	//output readSetConfigurer we must also output readSetter.
	if outputReadSetConfigurer {
		outputReadSetter = true
	}

	fields := structFields(location, s, allStructs)

	return &readerGenerator{
		s:                       s,
		outputReader:            outputReader,
		outputReadSetter:        outputReadSetter,
		outputReadSetConfigurer: outputReadSetConfigurer,
		fields:                  fields,
	}

}

//Output returns the code to append to the output for this struct.
func (r *readerGenerator) Output() string {
	var output string

	output += r.headerForStruct()

	if r.outputReader {
		output += templateOutput(readerTemplate, r.baseReaderGeneratorTemplateArguments())
	}
	if r.outputReadSetter {
		output += templateOutput(readSetterTemplate, r.baseReaderGeneratorTemplateArguments())
	}
	if r.outputReadSetConfigurer {
		output += templateOutput(readSetConfigurerTemplate, r.baseReaderGeneratorTemplateArguments())
	}
	return output
}

//baseReaderGeneratorTemplateArguments are base arguments that are passed to
//each template for readergeneratorm that is specific to a struct. Designed to
//be embedded anonymously in other structs passed to templates.
type baseReaderGeneratorTemplateArguments struct {
	StructName              string
	FirstLetter             string
	ReaderName              string
	OutputReadSetter        bool
	OutputReadSetConfigurer bool
}

func (r *readerGenerator) baseReaderGeneratorTemplateArguments() baseReaderGeneratorTemplateArguments {
	structName := r.s.Name
	//The prefix used to be "__" but that didn't lint correctly, so instead use
	//a non-latin prefix character that is like an a but with a dot (to make it
	//less likely to show up in autocompletes in IDEs)
	readerName := "ȧutoGenerated" + strings.Title(structName) + "Reader"

	return baseReaderGeneratorTemplateArguments{
		StructName:              structName,
		FirstLetter:             strings.ToLower(structName[:1]),
		ReaderName:              readerName,
		OutputReadSetter:        r.outputReadSetter,
		OutputReadSetConfigurer: r.outputReadSetConfigurer,
	}
}

func (r *readerGenerator) headerForStruct() string {

	//TODO: memoize propertyTypes/setterPropertyTypes because they don't
	//change within a run of this program.

	output := templateOutput(structHeaderTemplate,
		struct {
			baseReaderGeneratorTemplateArguments
			PropertyTypes []propertyType
			Fields        fieldsInfo
		}{
			baseReaderGeneratorTemplateArguments: r.baseReaderGeneratorTemplateArguments(),
			PropertyTypes:                        allValidTypes,
			Fields:                               r.fields,
		})

	for _, propType := range allValidTypes {

		var fieldsForType []fieldInfo

		for key, field := range r.fields {
			if field.Type == propType {
				fieldsForType = append(fieldsForType, fieldInfo{
					Name:    key,
					Mutable: field.Mutable,
					SubType: field.SubType,
				})
			}
		}

		sort.Slice(fieldsForType, func(i, j int) bool {
			return fieldsForType[i].Name < fieldsForType[j].Name
		})

		output += templateOutput(typedPropertyTemplate,
			struct {
				baseReaderGeneratorTemplateArguments
				PropType      propertyType
				FieldsForType []fieldInfo
			}{
				baseReaderGeneratorTemplateArguments: r.baseReaderGeneratorTemplateArguments(),
				PropType:                             propType,
				FieldsForType:                        fieldsForType,
			})
	}

	return output

}
