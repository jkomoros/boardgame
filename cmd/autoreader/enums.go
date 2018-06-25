package main

import (
	"errors"
	"github.com/abcum/lcp"
	enumpkg "github.com/jkomoros/boardgame/enum"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var displayNameRegExp = regexp.MustCompile(`display:\"(.*)\"`)
var transformUpperRegExp = regexp.MustCompile(`(?i)transform:\s*upper`)
var transformLowerRegExp = regexp.MustCompile(`(?i)transform:\s*lower`)
var transformNoneRegExp = regexp.MustCompile(`(?i)transform:\s*none`)

var enumHeaderTemplate *template.Template
var enumDelegateTemplate *template.Template
var enumItemTemplate *template.Template

func firstLetter(in string) string {

	if in == "" {
		return ""
	}

	return strings.ToLower(in[:1])
}

func init() {

	funcMap := template.FuncMap{
		"firstLetter": firstLetter,
	}

	enumHeaderTemplate = template.Must(template.New("enumheader").Funcs(funcMap).Parse(enumHeaderTemplateText))
	enumDelegateTemplate = template.Must(template.New("enumdelegate").Funcs(funcMap).Parse(enumDelegateTemplateText))
	enumItemTemplate = template.Must(template.New("enumitem").Parse(enumItemTemplateText))

}

type transform int

const (
	transformNone transform = iota
	transformUpper
	transformLower
)

type enum struct {
	//TODO: make most of thse fields private
	PackageName string
	Keys        []string
	//When BakeStringValues() is called, we take Transform, DefaultTransform,
	//and OverrideDisplayName and make the string values.
	bakedStringValues map[string]string
	//OverrideDisplayName contains a map of the Value string to override
	//value, if it exists. If it is in the map with value "" then it has been
	//overridden to have that value. If it is not in the map then it should be
	//default.
	OverrideDisplayName map[string]string
	Parents             map[string]string
	Transform           map[string]transform
	DefaultTransform    transform
	cachedPrefix        string
}

//findDelegateName looks through the given package to find the name of the
//struct that appears to represent the gameDelegate type, and returns its name.
func findDelegateName(packageASTs map[string]*ast.Package) ([]string, error) {

	var result []string

	for _, theAST := range packageASTs {
		for _, file := range theAST.Files {
			for _, decl := range file.Decls {

				//We're looking for function declarations like func (g
				//*gameDelegate) ConfigureMoves()
				//*boardgame.MoveTypeConfigBundle.

				funDecl, ok := decl.(*ast.FuncDecl)

				//Guess this decl wasn't a fun.
				if !ok {
					continue
				}

				if funDecl.Name.Name != "ConfigureMoves" {
					continue
				}

				if funDecl.Type.Params.NumFields() != 0 {
					continue
				}

				if funDecl.Type.Results.NumFields() != 1 {
					continue
				}

				returnFieldStar, ok := funDecl.Type.Results.List[0].Type.(*ast.StarExpr)

				if !ok {
					//OK, doesn't return a pointer, can't be a match.
					continue
				}

				returnFieldSelector, ok := returnFieldStar.X.(*ast.SelectorExpr)

				if !ok {
					//OK, there's no boardgame...
					continue
				}

				if returnFieldSelector.Sel.Name != "MoveTypeConfigBundle" {
					continue
				}

				returnFieldSelectorPackage, ok := returnFieldSelector.X.(*ast.Ident)

				if !ok {
					continue
				}

				if returnFieldSelectorPackage.Name != "boardgame" {
					continue
				}

				//TODO: verify the one return type is boardgame.MoveTypeConfigBundle

				if funDecl.Recv == nil || funDecl.Recv.NumFields() != 1 {
					//Verify i
					continue
				}

				//OK, it appears to be the right method. Extract out information about it.

				starExp, ok := funDecl.Recv.List[0].Type.(*ast.StarExpr)

				if !ok {
					return nil, errors.New("Couldn't cast candidate to star exp")
				}

				ident, ok := starExp.X.(*ast.Ident)

				if !ok {
					return nil, errors.New("Rest of star expression wasn't an ident")
				}

				result = append(result, ident.Name)

			}
		}
	}

	return result, nil
}

//filterDelegateNames takes delegate names we may want to export, and filters
//out any that already have a ConfigureEnums outputted.
func filterDelegateNames(candidates []string, packageASTs map[string]*ast.Package) []string {

	candidateMap := make(map[string]bool, len(candidates))

	for _, candidate := range candidates {
		candidateMap[candidate] = true
	}

	//Look through packageASTs and set to false any that we find a ConfigureEnums for.

	for _, theAST := range packageASTs {
		for _, file := range theAST.Files {

			//If the file was auto-generated by auto-enum (which by default is
			//at auto_enum.go but could be anywhere) then those definitions
			//don't count as manual definitions.
			if len(file.Comments) > 0 && strings.Contains(file.Comments[0].Text(), "It was generated by autoreader.") {
				continue
			}

			for _, decl := range file.Decls {

				//We're looking for function declarations like func (g
				//*gameDelegate) ConfigureMoves()
				//*boardgame.MoveTypeConfigBundle.

				funDecl, ok := decl.(*ast.FuncDecl)

				//Guess this decl wasn't a fun.
				if !ok {
					continue
				}

				if funDecl.Name.Name != "ConfigureEnums" {
					continue
				}

				if funDecl.Type.Params.NumFields() != 0 {
					continue
				}

				if funDecl.Type.Results.NumFields() != 1 {
					continue
				}

				returnFieldStar, ok := funDecl.Type.Results.List[0].Type.(*ast.StarExpr)

				if !ok {
					//OK, doesn't return a pointer, can't be a match.
					continue
				}

				returnFieldSelector, ok := returnFieldStar.X.(*ast.SelectorExpr)

				if !ok {
					//OK, there's no boardgame...
					continue
				}

				if returnFieldSelector.Sel.Name != "Set" {
					continue
				}

				returnFieldSelectorPackage, ok := returnFieldSelector.X.(*ast.Ident)

				if !ok {
					continue
				}

				if returnFieldSelectorPackage.Name != "enum" {
					continue
				}

				if funDecl.Recv == nil || funDecl.Recv.NumFields() != 1 {
					//Verify i
					continue
				}

				//OK, it appears to be the right method. Extract out information about it.

				starExp, ok := funDecl.Recv.List[0].Type.(*ast.StarExpr)

				if !ok {
					//Not expected, but whatever, it's safe to just include it
					continue
				}

				ident, ok := starExp.X.(*ast.Ident)

				if !ok {
					//Not expected, but whatever, it's safe to just include it
					continue
				}

				//If that struct type were one of the things we would export,
				//then note not to export it. If it wasn't already in, it
				//doesn't hurt to affirmatively say not to export it.
				candidateMap[ident.Name] = false

			}
		}
	}

	var result []string

	for name, include := range candidateMap {
		if !include {
			continue
		}
		result = append(result, name)
	}

	return result

}

func newEnum(packageName string, defaultTransform transform) *enum {
	return &enum{
		PackageName:         packageName,
		OverrideDisplayName: make(map[string]string),
		Transform:           make(map[string]transform),
		DefaultTransform:    defaultTransform,
	}
}

//findEnums processes the package at packageName and returns a list of enums
//that should be processed (that is, they have the magic comment)
func findEnums(packageASTs map[string]*ast.Package) (enums []*enum, err error) {

	for packageName, theAST := range packageASTs {
		for _, file := range theAST.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)

				if !ok {
					//Guess it wasn't a genDecl at all.
					continue
				}

				if genDecl.Tok != token.CONST {
					//We're only interested in Const decls.
					continue
				}

				if !enumConfig(genDecl.Doc.Text()) {
					//Must not have found the magic comment in the docs.
					continue
				}

				defaultTransform := configTransform(genDecl.Doc.Text(), transformNone)

				theEnum := newEnum(packageName, defaultTransform)

				for _, spec := range genDecl.Specs {

					valueSpec, ok := spec.(*ast.ValueSpec)

					if !ok {
						//Guess it wasn't a valueSpec after all!
						continue
					}

					if len(valueSpec.Names) != 1 {
						return nil, errors.New("Found an enum that had more than one name on a line. That's not allowed for now.")
					}

					keyName := valueSpec.Names[0].Name

					hasOverride, displayName := overrideDisplayname(valueSpec.Doc.Text())

					transform := configTransform(valueSpec.Doc.Text(), defaultTransform)

					theEnum.AddTransformKey(keyName, hasOverride, displayName, transform)

				}

				if len(theEnum.Keys) > 0 {
					enums = append(enums, theEnum)
				}

			}
		}
	}

	return enums, nil
}

var spaceReducer *regexp.Regexp
var titleCaseReplacer *strings.Replacer

//titleCaseToWords writes "ATitleCaseString" to "A Title Case String"
func titleCaseToWords(in string) string {

	//substantially recreated in moves/base.go

	if titleCaseReplacer == nil {

		var replacements []string

		for r := 'A'; r <= 'Z'; r++ {
			str := string(r)
			replacements = append(replacements, str)
			replacements = append(replacements, " "+str)
		}

		titleCaseReplacer = strings.NewReplacer(replacements...)

		spaceReducer = regexp.MustCompile(`\s+`)

	}

	titleCaseSplit := titleCaseReplacer.Replace(in)
	reducedSpaces := spaceReducer.ReplaceAllString(titleCaseSplit, " ")

	return strings.TrimSpace(reducedSpaces)

}

func processEnums(packageName string) (enumOutput string, err error) {

	packageASTs, err := parser.ParseDir(token.NewFileSet(), packageName, nil, parser.ParseComments)

	if err != nil {
		return "", errors.New("Parse error: " + err.Error())
	}

	enums, err := findEnums(packageASTs)

	if err != nil {
		return "", errors.New("Couldn't parse for enums: " + err.Error())
	}

	if len(enums) == 0 {
		//No enums. That's totally legit.
		return "", nil
	}

	delegateNames, err := findDelegateName(packageASTs)

	if err != nil {
		return "", errors.New("Failed to find delegate name: " + err.Error())
	}

	filteredDelegateNames := filterDelegateNames(delegateNames, packageASTs)

	output := enumHeaderForPackage(enums[0].PackageName, filteredDelegateNames)

	for i, e := range enums {
		if err := e.Legal(); err != nil {
			return "", errors.New(strconv.Itoa(i) + " enum had error: " + err.Error())
		}

		if err := e.BakeStringValues(); err != nil {
			return "", errors.New(strconv.Itoa(i) + " enum could not be baked: " + err.Error())
		}

		output += e.Output()

	}

	return output, nil

}

func enumConfig(docLines string) bool {

	for _, docLine := range strings.Split(docLines, "\n") {
		docLine = strings.ToLower(docLine)
		docLine = strings.TrimPrefix(docLine, "//")
		docLine = strings.TrimSpace(docLine)
		if strings.HasPrefix(docLine, magicDocLinePrefix) {
			return true
		}
	}

	return false
}

func configTransform(docLines string, defaultTransform transform) transform {
	for _, line := range strings.Split(docLines, "\n") {
		if transformLowerRegExp.MatchString(line) {
			return transformLower
		}
		if transformUpperRegExp.MatchString(line) {
			return transformUpper
		}
		if transformNoneRegExp.MatchString(line) {
			return transformNone
		}
	}

	return defaultTransform
}

func overrideDisplayname(docLines string) (hasOverride bool, displayName string) {
	for _, line := range strings.Split(docLines, "\n") {
		result := displayNameRegExp.FindStringSubmatch(line)

		if len(result) == 0 {
			continue
		}

		if len(result[0]) == 0 {
			continue
		}
		if len(result) != 2 {
			continue
		}

		//Found it! Even if the matched expression is "", that's fine. if
		//there are quoted strings that's fine, because that's exactly how
		//they should be output at the end.
		return true, result[1]

	}

	return false, ""
}

//BakeStringValues takes Key, Transform, DefaultTransform,
//OverrideDisplayValue and converts to a baked string value. Baked() must be
//false.
func (e *enum) BakeStringValues() error {
	if e.bakedStringValues != nil {
		return errors.New("String values already baked")
	}

	//Don't set field on struct yet, because e.Baked() shoudln't return true
	//unti lwe 're done, so StringValue will calculate what it should be live.
	bakedStringValues := make(map[string]string, len(e.Keys))

	for _, key := range e.Keys {
		bakedStringValues[key] = e.StringValue(key)
	}

	e.OverrideDisplayName = nil
	e.DefaultTransform = transformNone
	e.Transform = nil

	//Make sur eprefix is cached
	e.Prefix()

	e.bakedStringValues = bakedStringValues

	return nil
}

//Baked returnst true if BakeStringValues has been called.
func (e *enum) Baked() bool {
	return e.bakedStringValues != nil
}

//AddTransformKey adds a key to an enum that hasn't been baked yet.
func (e *enum) AddTransformKey(key string, overrideDisplay bool, overrideDisplayName string, transform transform) error {

	if e.Baked() {
		return errors.New("Can't add transform key to a baked enum")
	}

	//TODO: check for existing key name
	e.Keys = append(e.Keys, key)

	if overrideDisplay {
		e.OverrideDisplayName[key] = overrideDisplayName
	}

	e.Transform[key] = transform

	return nil
}

func (e *enum) AddBakedKey(key string, val string) error {
	if !e.Baked() {
		return errors.New("Can't add baked key to a non-baked enum")
	}

	//TODO: fail if prefix doesn't match

	//TODO: check for existing key name
	e.Keys = append(e.Keys, key)

	e.bakedStringValues[key] = val

	return nil
}

func (e *enum) RenameKey(oldKey, newKey string) {

	keyIndex := -1

	for i, key := range e.Keys {
		if key == oldKey {
			keyIndex = i
		}
	}

	if keyIndex < 0 {
		return
	}

	e.Keys[keyIndex] = newKey

	if e.Baked() {
		e.bakedStringValues[newKey] = e.bakedStringValues[oldKey]
		delete(e.bakedStringValues, oldKey)
		return
	}

	if _, ok := e.OverrideDisplayName[oldKey]; ok {
		e.OverrideDisplayName[newKey] = e.OverrideDisplayName[oldKey]
		delete(e.OverrideDisplayName, oldKey)
	}

	e.Transform[newKey] = e.Transform[oldKey]
	delete(e.Transform, oldKey)

}

func (e *enum) HasKey(key string) bool {
	for _, theKey := range e.Keys {
		if key == theKey {
			return true
		}
	}
	return false
}

//Output is the text to put into the final output in auto_enum.go
func (e *enum) Output() string {

	modifiedValues, parents := e.createParents()

	return e.baseOutput(e.Prefix(), modifiedValues, parents)

}

func (e *enum) ValueMap() map[string]string {
	//TODO: only regenerate this if a key or displayname has changed.
	result := make(map[string]string, len(e.Keys))
	for _, key := range e.Keys {
		result[key] = e.StringValue(key)
	}
	return result
}

//StringValue does all of the calulations and returns final value
func (e *enum) StringValue(key string) string {

	if e.bakedStringValues != nil {
		return e.bakedStringValues[key]
	}

	displayName, ok := e.OverrideDisplayName[key]

	if ok {
		return displayName
	}

	prefix := e.Prefix()

	withNoPrefix := strings.Replace(key, prefix, "", -1)
	expandedDelimiter := strings.Replace(withNoPrefix, "_", enumpkg.TREE_NODE_DELIMITER, -1)

	displayName = titleCaseToWords(expandedDelimiter)

	switch e.Transform[key] {
	case transformLower:
		displayName = strings.ToLower(displayName)
	case transformUpper:
		displayName = strings.ToUpper(displayName)
	}

	return displayName

}

//TreeEnum is whether or not we should output a TreeEnum.
func (e *enum) TreeEnum() bool {
	key := e.Prefix()
	if !e.HasKey(key) {
		return false
	}
	return e.StringValue(key) == ""
}

func (e *enum) PublicKeys() []string {
	//TODO: why do we have this and use it instead of Keys?
	var literals []string
	for _, literal := range e.Keys {
		if !fieldNamePublic(literal) {
			continue
		}
		literals = append(literals, literal)
	}
	return literals
}

func (e *enum) Prefix() string {

	if e.Baked() {
		//If baked, prefix has been explicitly set
		return e.cachedPrefix
	}

	//TODO: allow a cachedPrefix of ""
	if e.cachedPrefix != "" {
		return e.cachedPrefix
	}

	literals := e.PublicKeys()

	byteLiterals := make([][]byte, len(literals))

	for i, literal := range literals {
		byteLiterals[i] = []byte(literal)
	}

	if len(literals) == 0 {
		return ""
	}

	e.cachedPrefix = string(lcp.LCP(byteLiterals...))

	return e.cachedPrefix

}

//Legal will return an error if the enum isn't legal and shouldn't be output.
func (e *enum) Legal() error {

	if len(e.PublicKeys()) == 0 {
		return errors.New("No public keys")
	}

	if e.Prefix() == "" {
		return errors.New("Enum didn't have a shared prefix")
	}

	return nil

}

func enumHeaderForPackage(packageName string, delegateNames []string) string {

	output := templateOutput(enumHeaderTemplate, map[string]interface{}{
		"packageName": packageName,
	})

	//Ensure  a consistent ordering.
	sort.Strings(delegateNames)

	for _, delegateName := range delegateNames {
		output += templateOutput(enumDelegateTemplate, map[string]interface{}{
			"delegateName": delegateName,
		})
	}

	return output
}

//createMissingNodes returns a map like values, but with the addition of nodes
//who are implied in that they have children but are missing.
func createMissingNodes(values map[string]string) map[string]string {

	result := make(map[string]string, len(values))

	index := valueMapIndex(values)

	//We'll work up from the extremes.
	nextConstant := math.MinInt64

	for key, value := range values {

		//Copy over into result
		result[key] = value

		splitValue := strings.Split(value, enumpkg.TREE_NODE_DELIMITER)

		for i := 1; i < len(splitValue); i++ {
			joinedSubSet := strings.Join(splitValue[0:i], enumpkg.TREE_NODE_DELIMITER)

			//Check to make sure that has an entry in the map.
			if _, ok := index[joinedSubSet]; ok {
				//There was one, we're good.
				continue
			}

			//There wasn't one, need to create it.
			newKey := strconv.Itoa(nextConstant)
			nextConstant++
			newValue := joinedSubSet

			result[newKey] = newValue
			index[newValue] = newKey

		}

	}

	return result

}

//valueMapIndex takes a value map and returns an index that flips the mapping.
func valueMapIndex(values map[string]string) map[string]string {
	//We'll oftne need to reverse from original string value back up to the
	//constant key, so process that now.
	valueReverseMap := make(map[string]string, len(values))

	for key, val := range values {
		valueReverseMap[val] = key
	}

	return valueReverseMap
}

func (e *enum) createParents() (modifiedValues map[string]string, parents map[string]string) {

	//TODO: rename to reduceParents(), which creates the parent map TODO: add
	//a reduceNodeNames which takes the fully qualified name like " > " and
	//reduces to only be the last bit of name.

	//TODO: only allow this destructive processing to happen once on a given enum. (set a bit?)

	values := e.ValueMap()

	if !e.TreeEnum() {
		return values, nil
	}

	values = createMissingNodes(values)

	valueReverseMap := valueMapIndex(values)

	newValues := make(map[string]string, len(values))
	parentsResult := make(map[string]string, len(values))

	//Reduce the names of the resulting values to the last name in the hierachy

	for key, value := range values {

		splitValue := strings.Split(value, enumpkg.TREE_NODE_DELIMITER)

		lastValueComponent := splitValue[len(splitValue)-1]

		newValues[key] = lastValueComponent

	}

	//Set parents
	for key, value := range values {

		splitValue := strings.Split(value, enumpkg.TREE_NODE_DELIMITER)

		parentNode := valueReverseMap[""]

		if len(splitValue) >= 2 {
			//Not a node who points to root
			parentValue := strings.Join(splitValue[0:len(splitValue)-1], enumpkg.TREE_NODE_DELIMITER)
			parentNode = valueReverseMap[parentValue]
		}

		parentsResult[key] = parentNode
	}

	return newValues, parentsResult

}

func (e *enum) baseOutput(prefix string, values map[string]string, parents map[string]string) string {
	return templateOutput(enumItemTemplate, map[string]interface{}{
		"prefix":  prefix,
		"values":  values,
		"parents": parents,
	})
}

const enumHeaderTemplateText = `/************************************
 *
 * This file contains auto-generated methods to help configure enums. 
 * It was generated by autoreader.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package {{.packageName}}

import (
	"github.com/jkomoros/boardgame/enum"
)

var Enums = enum.NewSet()

`

const enumDelegateTemplateText = `//ConfigureEnums simply returns Enums, the auto-generated Enums variable. This
//is output because {{.delegateName}} appears to be a struct that implements
//boardgame.GameDelegate, and does not already have a ConfigureEnums
//explicitly defined.
func ({{firstLetter .delegateName}} *{{.delegateName}}) ConfigureEnums() *enum.Set {
	return Enums
}

`

const enumItemTemplateText = `var {{.prefix}}Enum = Enums.MustAdd{{if .parents}}Tree{{end}}("{{.prefix}}", map[int]string{
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
