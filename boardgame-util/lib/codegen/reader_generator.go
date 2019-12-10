package codegen

import (
	"sort"
	"strings"

	"github.com/MarcGrol/golangAnnotations/model"
	"github.com/jkomoros/boardgame"
)

//allValidTypes is an enumeration of all types in order.
var allValidTypes []boardgame.PropertyType

//highestProperty is the highest enum in the PropertyType enum.
const highestProperty = boardgame.TypeTimer

func init() {
	//Only need space for highestProperty because we skip TypeIllegal.
	allValidTypes = make([]boardgame.PropertyType, highestProperty)
	//We skip TypeIllegal
	for i := 0; i < int(highestProperty); i++ {
		allValidTypes[i] = boardgame.PropertyType(i + 1)
	}
}

//readerGenerator represents a strucxt in the imported code that had the magic
//codegen tag attached, meaning that we should generate code for it.
type readerGenerator struct {
	s                       model.Struct
	outputReader            bool
	outputReadSetter        bool
	outputReadSetConfigurer bool
	//TODO: pop all of this directly into the struct
	fields *typeInfo
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

	//propertyTypes is short name, golangValue
	propertyTypes := make(map[string]string)
	setterPropertyTypes := make(map[string]string)

	for i := boardgame.TypeInt; i <= boardgame.TypeTimer; i++ {

		key := strings.TrimPrefix(i.String(), "Type")

		goLangType := key
		setterGoLangType := ""
		switch key {
		case "Bool":
			goLangType = "bool"
		case "Int":
			goLangType = "int"
		case "String":
			goLangType = "string"
		case "PlayerIndex":
			goLangType = "boardgame.PlayerIndex"
		case "IntSlice":
			goLangType = "[]int"
		case "BoolSlice":
			goLangType = "[]bool"
		case "StringSlice":
			goLangType = "[]string"
		case "PlayerIndexSlice":
			goLangType = "[]boardgame.PlayerIndex"
		case "Enum":
			goLangType = "enum.ImmutableVal"
			setterGoLangType = "enum.Val"
		case "Stack":
			goLangType = "boardgame.ImmutableStack"
			setterGoLangType = "boardgame.Stack"
		case "Board":
			goLangType = "boardgame.ImmutableBoard"
			setterGoLangType = "boardgame.Board"
		case "Timer":
			goLangType = "boardgame.ImmutableTimer"
			setterGoLangType = "boardgame.Timer"
		default:
			goLangType = "UNKNOWN"
		}

		if setterGoLangType == "" {
			setterGoLangType = goLangType
		}

		propertyTypes[key] = goLangType
		setterPropertyTypes[key] = setterGoLangType
	}

	output := templateOutput(structHeaderTemplate,
		struct {
			baseReaderGeneratorTemplateArguments
			PropertyTypes       map[string]string
			SetterPropertyTypes map[string]string
			Fields              *typeInfo
		}{
			baseReaderGeneratorTemplateArguments: r.baseReaderGeneratorTemplateArguments(),
			PropertyTypes:                        propertyTypes,
			SetterPropertyTypes:                  setterPropertyTypes,
			Fields:                               r.fields,
		})

	for _, propType := range allValidTypes {

		var namesForType []nameForTypeInfo

		for key, val := range r.fields.Types {
			if val == propType {
				namesForType = append(namesForType, nameForTypeInfo{
					Name:        key,
					Mutable:     r.fields.Mutable[key],
					UpConverter: r.fields.UpConverter[key],
				})
			}
		}

		sort.Slice(namesForType, func(i, j int) bool {
			return namesForType[i].Name < namesForType[j].Name
		})

		output += templateOutput(typedPropertyTemplate,
			struct {
				baseReaderGeneratorTemplateArguments
				PropType     boardgame.PropertyType
				NamesForType []nameForTypeInfo
			}{
				baseReaderGeneratorTemplateArguments: r.baseReaderGeneratorTemplateArguments(),
				PropType:                             propType,
				NamesForType:                         namesForType,
			})
	}

	return output

}
