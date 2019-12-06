package codegen

import "github.com/MarcGrol/golangAnnotations/model"

//newReaderGenerator processes the given struct and then outputs a generator if
//any code is necessary to be output.
func newReaderGenerator(s model.Struct, location string, allStructs []model.Struct) *readerGenerator {
	outputReader, outputReadSetter, outputReadSetConfigurer := structConfig(s.DocLines)

	if !outputReader && !outputReadSetter && !outputReadSetConfigurer {
		return nil
	}

	types := structTypes(location, s, allStructs)

	return &readerGenerator{
		s:                       s,
		outputReader:            outputReader,
		outputReadSetter:        outputReadSetter,
		outputReadSetConfigurer: outputReadSetConfigurer,
		types:                   types,
	}

}

//Output returns the code to append to the output for this struct.
func (r *readerGenerator) Output() string {
	var output string

	output += headerForStruct(r.s.Name, r.types, r.outputReadSetter, r.outputReadSetConfigurer)

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

//readerGenerator represents a strucxt in the imported code that had the magic
//codegen tag attached, meaning that we should generate code for it.
type readerGenerator struct {
	s                       model.Struct
	outputReader            bool
	outputReadSetter        bool
	outputReadSetConfigurer bool
	//TODO: pop all of this directly into the struct
	types *typeInfo
}
