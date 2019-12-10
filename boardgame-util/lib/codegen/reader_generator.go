package codegen

import (
	"sort"
	"strings"

	"github.com/MarcGrol/golangAnnotations/model"
	"github.com/jkomoros/boardgame"
)

//allTypes is an enumeration of all types in order.
var allTypes []boardgame.PropertyType

//highestProperty is the highest enum in the PropertyType enum.
const highestProperty = boardgame.TypeTimer

func init() {
	allTypes = make([]boardgame.PropertyType, highestProperty+1)
	for i := 0; i <= int(highestProperty); i++ {
		allTypes[i] = boardgame.PropertyType(i)
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
		output += readerForStruct(r.s.Name)
	}
	if r.outputReadSetter {
		output += readSetterForStruct(r.s.Name)
	}
	if r.outputReadSetConfigurer {
		output += readSetConfigurerForStruct(r.s.Name)
	}
	return output
}

func (r *readerGenerator) headerForStruct() string {

	//TODO: memoize propertyTypes/setterPropertyTypes because they don't
	//change within a run of this program.

	structName := r.s.Name

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
			StructName              string
			FirstLetter             string
			ReaderName              string
			PropertyTypes           map[string]string
			SetterPropertyTypes     map[string]string
			Fields                  *typeInfo
			OutputReadSetter        bool
			OutputReadSetConfigurer bool
		}{
			StructName:              structName,
			FirstLetter:             strings.ToLower(structName[:1]),
			ReaderName:              readerStructName(structName),
			PropertyTypes:           propertyTypes,
			SetterPropertyTypes:     setterPropertyTypes,
			Fields:                  r.fields,
			OutputReadSetter:        r.outputReadSetter,
			OutputReadSetConfigurer: r.outputReadSetConfigurer,
		})

	sortedKeys := make([]string, len(propertyTypes))
	i := 0

	for propType := range propertyTypes {
		sortedKeys[i] = propType
		i++
	}

	sort.Strings(sortedKeys)

	for _, propType := range sortedKeys {

		goLangType := propertyTypes[propType]

		zeroValue := "nil"

		switch propType {
		case "Bool":
			zeroValue = "false"
		case "Int":
			zeroValue = "0"
		case "String":
			zeroValue = "\"\""
		case "PlayerIndex":
			zeroValue = "0"
		case "IntSlice":
			zeroValue = "[]int{}"
		case "BoolSlice":
			zeroValue = "[]bool{}"
		case "StringSlice":
			zeroValue = "[]string{}"
		case "PlayerIndexSlice":
			zeroValue = "[]boardgame.PlayerIndex{}"
		}

		var namesForType []nameForTypeInfo

		for key, val := range r.fields.Types {
			if val.String() == "Type"+propType {
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

		setterPropType := propType

		outputMutableGetter := false

		switch propType {
		case "Enum":
			setterPropType = "Enum"
			outputMutableGetter = true
		case "Stack":
			setterPropType = "Stack"
			outputMutableGetter = true
		case "Board":
			setterPropType = "Board"
			outputMutableGetter = true
		case "Timer":
			setterPropType = "Timer"
			outputMutableGetter = true
		}

		setterGoLangType := setterPropertyTypes[setterPropType]

		output += templateOutput(typedPropertyTemplate, map[string]interface{}{
			"structName":              structName,
			"firstLetter":             strings.ToLower(structName[:1]),
			"readerName":              readerStructName(structName),
			"propType":                propType,
			"setterPropType":          setterPropType,
			"namesForType":            namesForType,
			"goLangType":              goLangType,
			"setterGoLangType":        setterGoLangType,
			"outputMutableGetter":     outputMutableGetter,
			"zeroValue":               zeroValue,
			"outputReadSetter":        r.outputReadSetter,
			"outputReadSetConfigurer": r.outputReadSetConfigurer,
		})
	}

	return output

}
