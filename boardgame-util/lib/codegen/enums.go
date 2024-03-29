package codegen

import (
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/abcum/lcp"
	enumpkg "github.com/jkomoros/boardgame/enum"
)

var displayNameRegExp = regexp.MustCompile(`display:\"(.*)\"`)
var combineRegExp = regexp.MustCompile(`combine:\s*(.*)`)
var transformUpperRegExp = regexp.MustCompile(`(?i)transform:\s*upper`)
var transformLowerRegExp = regexp.MustCompile(`(?i)transform:\s*lower`)
var transformNoneRegExp = regexp.MustCompile(`(?i)transform:\s*none`)

type transform int

//This used to be "_" which was clearer, but then the constant names would have
//"_" in them, which golint doesn't like. It was going to be "0" but that
//legitimately shows up in some cases, like in components/playingcards/Rank10.
//So instead use "010" which, because of the leading "0" is way less likely to
//show up.
const explicitTreeCharacter = "010"
const oldExplicitTreeCharacter = "_"

const (
	transformNone transform = iota
	transformUpper
	transformLower
)

type enum struct {
	PackageName string
	keys        []string
	//When BakeStringValues() is called, we take Transform, DefaultTransform,
	//and OverrideDisplayName and make the string values.
	bakedStringValues map[string]string
	//newKeys is keys that need to be created (only implied)
	newKeys []string
	//OverrideDisplayName contains a map of the Value string to override
	//value, if it exists. If it is in the map with value "" then it has been
	//overridden to have that value. If it is not in the map then it should be
	//default.
	overrideDisplayName map[string]string
	transform           map[string]transform
	parents             map[string]string
	defaultTransform    transform
	cachedPrefix        string
	processed           bool
	//the name picked up from the combineRegExp, if one exists.
	combineName string
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
				//*boardgame.MoveConfigBundle.

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

				arrayType, ok := funDecl.Type.Results.List[0].Type.(*ast.ArrayType)

				if !ok {
					//OK, there's no [] so can't be right type
					continue
				}

				sel, ok := arrayType.Elt.(*ast.SelectorExpr)

				if !ok {
					//There's no 'boardgame'
					continue
				}

				if sel.Sel.Name != "MoveConfig" {
					continue
				}

				returnFieldSelectorPackage, ok := sel.X.(*ast.Ident)

				if !ok {
					continue
				}

				if returnFieldSelectorPackage.Name != "boardgame" {
					continue
				}

				//TODO: verify the one return type is boardgame.MoveConfigBundle

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
			if len(file.Comments) > 0 && strings.Contains(file.Comments[0].Text(), "It was generated by the codegen package via 'boardgame-util codegen'") {
				continue
			}

			for _, decl := range file.Decls {

				//We're looking for function declarations like func (g
				//*gameDelegate) ConfigureMoves()
				//*boardgame.MoveConfigBundle.

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
		overrideDisplayName: make(map[string]string),
		transform:           make(map[string]transform),
		defaultTransform:    defaultTransform,
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

				if !enumConfig(genDecl.Doc) {
					//Must not have found the magic comment in the docs.
					continue
				}

				defaultTransform := configTransform(genDecl.Doc, transformNone)

				theEnum := newEnum(packageName, defaultTransform)

				theEnum.combineName = configCombineName(genDecl.Doc)

				for _, spec := range genDecl.Specs {

					valueSpec, ok := spec.(*ast.ValueSpec)

					if !ok {
						//Guess it wasn't a valueSpec after all!
						continue
					}

					if len(valueSpec.Names) != 1 {
						return nil, errors.New("found an enum that had more than one name on a line. That's not allowed for now")
					}

					keyName := valueSpec.Names[0].Name

					hasOverride, displayName := overrideDisplayname(valueSpec.Doc.Text())

					transform := configTransform(valueSpec.Doc, defaultTransform)

					theEnum.AddTransformKey(keyName, hasOverride, displayName, transform)

				}

				if len(theEnum.Keys()) > 0 {
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

/*

ProcessEnums processes the given package and outputs the contents of a file
representing the auto-generated boilerplate for those enums. If it finds a
const() block at the top-level decorated with the magic comment
(boardgame:codegen) it will generate enum boilerplate. See the package doc of
enum for more on what you need to include.

auto-generated enums will automatically have values like prefixVeryLongName set
to have a string value of "Very Long Name"; that is title-case will be taken to
mean word boundaries. Enum values can be package-private (start with lower-case)
or package public, but the two different types will be considered parts of
different enums. If you want to transform the created values to lowercase or
uppercase, include a line of `transform:lower` or `transform:upper`,
respectively, in the comment lines immediately before the constant.
`transform:none` means default behavior, leave as title case. If you want to
change the default transform for an entire const group, have the transform line
in the comment block above the constant block.  If you want to override a
specific item in the enum's name, include a comment immediately above that
matches that pattern `display:"myVal"`, where myVal is the exact string to use.
myVal may be zero-length, and may include quoted quotes. If your enum has a key
that is named with the prefix of the rest of the enum values, and evaluates to
0, then a TreeEnum will be created. See the documentation in the enum package
for how to control nesting in a TreeEnum. If enums are autogenerated, and the
struct in your package that appears to be your gameDelegate doesn't already have
a ConfigureEnums(), one will be generated for you.

If your enum contains a comment in its documentation like `combine:Foo` then
set.Combine() will be additionally created, where each enum with that same
combine name is combined together. As a special case, if the name is 'group'
(the default group enum name), then boardgame.BaseGroupEnum will also be
included into the group. Remember that enums that are combined can't overlap in
any int or string keys. It's idiomatic to start the nth enum not with iota but
`lastEnumMaxVal + 1 + iota` to automatically ensure no overlap.

*/
func ProcessEnums(packageName string) (enumOutput string, err error) {

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

	groups := make(map[string][]*enum)

	for _, e := range enums {
		if e.combineName != "" {
			groups[e.combineName] = append(groups[e.combineName], e)
		}
	}

	output := enumHeaderForPackage(enums[0].PackageName, filteredDelegateNames)

	for i, e := range enums {

		if err := e.Process(); err != nil {
			return "", errors.New(strconv.Itoa(i) + " enum could not be processed: " + err.Error())
		}
		enumOutput, err := e.Output()
		if err != nil {
			return "", errors.New(strconv.Itoa(i) + " enum output failed: " + err.Error())
		}
		output += enumOutput

	}

	//ensure that groups are always output in the same order
	var groupNamesInOrder []string
	for name := range groups {
		groupNamesInOrder = append(groupNamesInOrder, name)
	}
	sort.Strings(groupNamesInOrder)

	for _, name := range groupNamesInOrder {
		group := groups[name]
		//TODO: if name is defaultGroupsName, then also emit
		//boardgame.BaseGroupEnum, combined with the extra boardgame import
		var varNames []string
		for _, item := range group {
			varNames = append(varNames, item.Prefix()+"Enum")
		}
		output += groupOutput(name, varNames)
	}

	formattedEnumBytes, err := format.Source([]byte(output))

	if err != nil {
		if debugSaveBadCode {
			formattedEnumBytes = []byte(enumOutput)
		} else {
			return "", errors.New("Couldn't go fmt code for enums: " + err.Error())
		}
	}

	return string(formattedEnumBytes), nil

}

func enumConfig(comment *ast.CommentGroup) bool {

	if comment == nil {
		return false
	}

	for _, line := range comment.List {
		docLine := line.Text
		docLine = strings.ToLower(docLine)
		docLine = strings.TrimPrefix(docLine, "//")
		docLine = strings.TrimSpace(docLine)
		if strings.HasPrefix(docLine, magicDocLinePrefix) {
			return true
		}
	}

	return false
}

func configTransform(comment *ast.CommentGroup, defaultTransform transform) transform {
	if comment == nil {
		return defaultTransform
	}
	for _, commentLine := range comment.List {
		line := commentLine.Text
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

func configCombineName(comment *ast.CommentGroup) string {

	if comment == nil {
		return ""
	}

	for _, line := range comment.List {
		result := combineRegExp.FindStringSubmatch(line.Text)

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
		return result[1]

	}

	return ""
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

//Process should be called after all items ahve been added. Does lots of
//processing.
func (e *enum) Process() error {

	if e.processed {
		return errors.New("already processed")
	}

	if err := e.Legal(); err != nil {
		return errors.New("Enum not legal: " + err.Error())
	}

	if err := e.bakeStringValues(); err != nil {
		return errors.New("Couldn't bake string values: " + err.Error())
	}

	if e.TreeEnum() {

		if err := e.autoAddDelimiters(); err != nil {
			return errors.New("Couldn't auto add delimiters: " + err.Error())
		}

		if err := e.createMissingParents(); err != nil {
			return errors.New("Couldn't make missing parents: " + err.Error())
		}

		if err := e.makeParents(); err != nil {
			return errors.New("Couldn't make parents: " + err.Error())
		}

		e.reduceNodeStringValues()

	}

	e.processed = true

	return nil
}

//bakeStringValues takes Key, Transform, DefaultTransform,
//OverrideDisplayValue and converts to a baked string value. Baked() must be
//false. Will fail if e.Legal() returns an error. Should only be called from within Process().
func (e *enum) bakeStringValues() error {

	if e.bakedStringValues != nil {
		return errors.New("String values already baked")
	}

	//Don't set field on struct yet, because e.Baked() shoudln't return true
	//unti lwe 're done, so StringValue will calculate what it should be live.
	bakedStringValues := make(map[string]string, len(e.Keys()))

	for _, key := range e.Keys() {
		bakedStringValues[key] = e.StringValue(key)
	}

	e.overrideDisplayName = nil
	e.defaultTransform = transformNone
	e.transform = nil

	//Make sur eprefix is cached
	e.Prefix()

	e.bakedStringValues = bakedStringValues

	return nil
}

//Baked returnst true if BakeStringValues has been called.
func (e *enum) baked() bool {
	return e.bakedStringValues != nil
}

//AddTransformKey adds a key to an enum that hasn't been baked yet.
func (e *enum) AddTransformKey(key string, overrideDisplay bool, overrideDisplayName string, transform transform) error {

	if e.baked() {
		return errors.New("Can't add transform key to a baked enum")
	}

	if e.HasKey(key) {
		return errors.New(key + " already exists")
	}

	e.keys = append(e.keys, key)

	if overrideDisplay {
		e.overrideDisplayName[key] = overrideDisplayName
	}

	e.transform[key] = transform

	return nil
}

//addBakedKey adds keys after bakeStringValues has been called. Should only be
//called between baking and being fully processed.
func (e *enum) addBakedKey(key string, val string) error {

	if e.processed {
		return errors.New("Can't add baked key to already rpocessed enum")
	}

	if !e.baked() {
		return errors.New("Can't add baked key to a non-baked enum")
	}

	if e.HasKey(key) {
		return errors.New(key + " already exists")
	}

	if !strings.HasPrefix(key, e.Prefix()) {
		if _, err := strconv.Atoi(key); err != nil {
			return errors.New("key must either have prefix " + e.Prefix() + " or be an int")
		}
	}

	e.keys = append(e.keys, key)
	e.newKeys = append(e.newKeys, key)

	e.bakedStringValues[key] = val

	return nil
}

//NewKeys returns a list of new keys that were implied in this tree enum but
//need to be explciitly created in auto_enum.
func (e *enum) NewKeys() []string {
	sort.Strings(e.newKeys)
	return e.newKeys
}

func (e *enum) HasKey(key string) bool {
	for _, theKey := range e.Keys() {
		if key == theKey {
			return true
		}
	}
	return false
}

func (e *enum) Parents() map[string]string {
	return e.parents
}

//Output is the text to put into the final output in auto_enum.go
func (e *enum) Output() (string, error) {

	if !e.processed {
		return "", errors.New("not processed. Call Process first")
	}

	return e.baseOutput(e.Prefix(), e.ValueMap(), e.Parents(), e.NewKeys()), nil

}

func (e *enum) ValueMap() map[string]string {
	//TODO: only regenerate this if a key or displayname has changed.
	result := make(map[string]string, len(e.Keys()))
	for _, key := range e.Keys() {
		result[key] = e.StringValue(key)
	}
	return result
}

func (e *enum) ReverseValueMap() map[string]string {
	//TODO: only regenerate this if a key or displayname has changed.
	result := make(map[string]string, len(e.Keys()))
	for _, key := range e.Keys() {
		result[e.StringValue(key)] = key
	}
	return result
}

//StringValue does all of the calulations and returns final value
func (e *enum) StringValue(key string) string {

	if e.bakedStringValues != nil {
		return e.bakedStringValues[key]
	}

	displayName, ok := e.overrideDisplayName[key]

	if ok {
		return displayName
	}

	prefix := e.Prefix()

	withNoPrefix := strings.Replace(key, prefix, "", 1)
	expandedDelimiter := strings.Replace(withNoPrefix, explicitTreeCharacter, enumpkg.TreeNodeDelimiter, -1)

	displayName = titleCaseToWords(expandedDelimiter)

	switch e.transform[key] {
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

func (e *enum) Keys() []string {
	return e.keys
}

func (e *enum) Prefix() string {

	if e.baked() {
		//If baked, prefix has been explicitly set, even if it's "".
		return e.cachedPrefix
	}

	if e.cachedPrefix != "" {
		return e.cachedPrefix
	}

	literals := e.Keys()

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

	if len(e.Keys()) == 0 {
		return errors.New("No public keys")
	}

	for _, key := range e.Keys() {
		if strings.Contains(key, oldExplicitTreeCharacter) {
			return errors.New("Key " + key + " had a '" + oldExplicitTreeCharacter + "' as an explicit delimiter, but that should be changed to '" + explicitTreeCharacter + "'")
		}
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

/*

PhaseOne -> "One" -> "One"
PhaseOneOne -> "One One" -> "One > One"
PhaseOneTwo -> "One Two" -> "One > Two"
PhaseNextOneOne -> "Next One One" -> "Next One > One"
PhaseNextOneTwo -> "Next One Two" -> "Next One > Two"
PhaseTwo010One -> "Two > One" -> "Two > One"
*/

type delimiterTree struct {
	parent          *delimiterTree
	children        map[string]*delimiterTree
	manuallyCreated bool
	//For value string that ends here, its value.
	terminalKey string
}

//addString goes through and adds addChild down the whole way. If it consumes
//a ">" off the front, then it does manuallyCreated = true. The last item has
//addChild with a terminalKey of the passed terminalKey.
func (t *delimiterTree) addString(names []string, terminalKey string, lastItemWasDelimiter bool) error {

	delimiter := strings.TrimSpace(enumpkg.TreeNodeDelimiter)

	if len(names) == 0 {
		return errors.New("addString called with no names")
	}

	if len(names) == 1 {
		return t.addChild(names[0], lastItemWasDelimiter, terminalKey)
	}

	if names[0] == delimiter {
		return t.addString(names[1:], terminalKey, true)
	}

	if err := t.addChild(names[0], lastItemWasDelimiter, ""); err != nil {
		return err
	}

	return t.children[names[0]].addString(names[1:], terminalKey, false)

}

//addChild adds a child node, if one doesn't already exist. manuallyCreated is
//always the OR of existing value on the ndoe. terminalKey should only be
//non-"" if this is literally the end of the string.
func (t *delimiterTree) addChild(name string, manuallyCreated bool, terminalKey string) error {

	var child *delimiterTree

	if name == "" {

		if t.parent != nil {
			return errors.New("Trying to set '' on non root node")
		}

		//Special case: setting the terminalKey for the root node
		child = t
	} else {

		child = t.children[name]

		if child == nil {
			child = &delimiterTree{
				parent:   t,
				children: make(map[string]*delimiterTree),
			}
			t.children[name] = child
		}

		child.manuallyCreated = child.manuallyCreated || manuallyCreated

	}

	if terminalKey != "" {
		if child.terminalKey != "" {
			return errors.New("Child already had terminalKey set, wanted to set again (" + name + " : " + terminalKey + " : " + child.terminalKey + " : " + t.value() + ")")
		}
		child.terminalKey = terminalKey
	}

	return nil
}

//elideSingleParents, if this node has only one child, and was not
//manuallyCreated, elides itself.
func (t *delimiterTree) elideSingleParents() error {

	if t.parent != nil {

		var child *delimiterTree
		for _, c := range t.children {
			child = c
		}

		if len(t.children) == 1 && !child.manuallyCreated {
			if err := t.mergeDown(); err != nil {
				return errors.New("Couldn't merge down: " + err.Error())
			}
		}
	}
	for _, child := range t.children {
		if err := child.elideSingleParents(); err != nil {
			return err
		}
	}

	return nil
}

//mergeDown is called when this node should merge downward to its child.
func (t *delimiterTree) mergeDown() error {

	if t.parent == nil {
		return errors.New("Can't merge down a root node")
	}

	parentKeyName := ""
	for key, val := range t.parent.children {
		if val == t {
			parentKeyName = key
		}
	}

	if parentKeyName == "" {
		return errors.New("Unexpectedly couldn't find self in parent")
	}

	if len(t.children) != 1 {
		return errors.New("Merging down only legal if have one child")
	}

	var name string
	var child *delimiterTree

	for n, c := range t.children {
		//There should be only one item; this is basically just fetching the
		//one item.
		name = n
		child = c
	}

	if child.manuallyCreated {
		return errors.New("Merging down not legal onto a manually created node")
	}

	newName := parentKeyName + " " + name

	//Elide us out of the chain

	//Point up to the grandparent
	child.parent = t.parent

	//Point from the parent down to the child
	delete(t.parent.children, parentKeyName)
	t.parent.children[newName] = child

	return nil

}

//value returns the string value by walking parents
func (t *delimiterTree) value() string {
	if t.parent == nil {
		return ""
	}

	nameInParent := ""

	for name, child := range t.parent.children {
		if child == t {
			nameInParent = name
			break
		}
	}

	if nameInParent == "" {
		return nameInParent
	}

	parentValue := t.parent.value()

	if parentValue == "" {
		return nameInParent
	}

	return parentValue + enumpkg.TreeNodeDelimiter + nameInParent
}

//keyValues returns the key -> value mapping encoded in this tree, recursively
func (t *delimiterTree) keyValues() map[string]string {

	result := make(map[string]string)

	//Base case if we're a terminal
	if t.terminalKey != "" {
		result[t.terminalKey] = t.value()
	}

	for _, child := range t.children {
		for key, val := range child.keyValues() {
			result[key] = val
		}
	}

	return result

}

//autoAddDelimiters should only be called by Process. It adds delimiters to
//string values at implied breaks.
func (e *enum) autoAddDelimiters() error {

	//The general approach to this is that we'll create a tree of nodes and
	//their terminal keys, splitting at every place there COULD be a delimiter
	//(that is, every place there is a an explicit " > " break or an implicit
	//one of a " "). We build up this tree keeping track of which breaks were
	//explicit (and thus should not be removed) and which ones are just
	//speculative (i.e. at a word break). Then, go through and reduce out ones
	//that shouldn't be there: that is, that have a single child and their
	//child wasn't explicitly created. Then re-derive the string values from
	//the tweaked tree.

	tree := &delimiterTree{
		children: make(map[string]*delimiterTree),
	}

	for key, value := range e.ValueMap() {

		//TODO: handle values that have been transformed differently: should
		//compare the same when creating tree, but key/values needs to be
		//rewritten back to proper case at end.

		splitValue := strings.Split(value, " ")

		if err := tree.addString(splitValue, key, false); err != nil {
			return errors.New("Couldn't add " + key + ", " + value + " to tree delimiter: " + err.Error())
		}
	}

	if err := tree.elideSingleParents(); err != nil {
		return errors.New("Couldn't elide single parents: " + err.Error())
	}

	//This step wouldn't work if we ahd down-cased all of the items.
	e.bakedStringValues = tree.keyValues()

	return nil

}

//createMissingParents should only be called within Process. Creates any
//parent nodes that are implied but not explicitly provided.
func (e *enum) createMissingParents() error {

	index := e.ReverseValueMap()

	for _, value := range e.ValueMap() {

		splitValue := strings.Split(value, enumpkg.TreeNodeDelimiter)

		for i := 1; i < len(splitValue); i++ {
			joinedSubSet := strings.Join(splitValue[0:i], enumpkg.TreeNodeDelimiter)

			//Check to make sure that has an entry in the map.
			if _, ok := index[joinedSubSet]; ok {
				//There was one, we're good.
				continue
			}

			//There wasn't one, need to create it.
			newKey := e.Prefix() + joinedSubSet
			newKey = strings.Replace(newKey, enumpkg.TreeNodeDelimiter, explicitTreeCharacter, -1)
			newKey = strings.Replace(newKey, " ", "", -1)
			//reduce "010" to "" if that's unambiguous
			newKey = e.reduceProposedKey(newKey)
			newValue := joinedSubSet

			if err := e.addBakedKey(newKey, newValue); err != nil {
				return errors.New("Couldn't add implied new key: " + err.Error())
			}
			index[newValue] = newKey

		}

	}

	return nil

}

//reduceNewKey is given a proposed key, like "PhaseBlueGreen0One". It returns
//a string that has as many of the "010" elided as makes sense. Currently this
//is done by just mimicking whatever the explicit constants do.
func (e *enum) reduceProposedKey(proposedKey string) string {

	for _, option := range reducedKeyPermutations(proposedKey) {
		for _, key := range e.Keys() {
			if strings.HasPrefix(key, option) {
				return option
			}
		}
	}

	//None of the other reductions worked, so just return the current one.
	return proposedKey
}

//reducedKeypermutations returns all possible versions of this key with 0 to n
//of the "010" replaced with "" (but does not return the proposedKey itself).
func reducedKeyPermutations(proposedKey string) []string {
	pieces := strings.Split(proposedKey, explicitTreeCharacter)
	if len(pieces) == 1 {
		//No "010", so no options to return
		return nil
	}
	var result []string

	for _, mask := range maskPermuations(len(pieces) - 1) {
		str := pieces[0]
		for i, b := range mask {
			if b {
				str += explicitTreeCharacter
			}
			str += pieces[i+1]
		}
		result = append(result, str)
	}

	return result

}

//maskPermutations returns all possible bitmasks of length k
func maskPermuations(k int) [][]bool {
	total := int(math.Pow(2, float64(k)))
	result := make([][]bool, total)

	lastItem := make([]bool, k)
	result[0] = lastItem

	for i := 1; i < total; i++ {
		item := make([]bool, k)
		copy(item, lastItem)
		//Flip the last item.
		item[k-1] = !item[k-1]
		for j := k - 1; j >= 0; j-- {
			//Is the item we just flipped now true? if so, no need to carry
			//over to the left, can stop now.
			if item[j] {
				break
			}
			//Carry over by flipping the next bit to the left.
			item[j-1] = !item[j-1]
		}

		result[i] = item
	}
	return result
}

//makeParents should only be called by e.Process(). It creates the parents relationship.
func (e *enum) makeParents() error {

	if e.parents != nil {
		return errors.New("Parents already created")
	}

	index := e.ReverseValueMap()

	e.parents = make(map[string]string, len(e.Keys()))

	//Set parents
	for key, value := range e.ValueMap() {

		splitValue := strings.Split(value, enumpkg.TreeNodeDelimiter)

		//default to parent being the root node
		parentNode := index[""]

		if len(splitValue) >= 2 {
			//Not a node who points to root
			parentValue := strings.Join(splitValue[0:len(splitValue)-1], enumpkg.TreeNodeDelimiter)
			parentNode = index[parentValue]
		}

		e.parents[key] = parentNode
	}

	return nil

}

//reduceNodeStringValues should only be called by e.Process(). Reduces the
//display name to be just the last bit of the name.
func (e *enum) reduceNodeStringValues() {

	for key, value := range e.ValueMap() {

		splitValue := strings.Split(value, enumpkg.TreeNodeDelimiter)

		lastValueComponent := splitValue[len(splitValue)-1]

		e.bakedStringValues[key] = lastValueComponent

	}

}

func (e *enum) baseOutput(prefix string, values map[string]string, parents map[string]string, newKeys []string) string {

	firstKey := ""

	if len(newKeys) > 0 {
		firstKey = newKeys[0]
		newKeys = newKeys[1:]
	}

	return templateOutput(enumItemTemplate, map[string]interface{}{
		"prefix":      prefix,
		"values":      values,
		"parents":     parents,
		"firstNewKey": firstKey,
		"restNewKeys": newKeys,
	})
}

func groupOutput(name string, enumVarNames []string) string {
	return templateOutput(enumGroupTemplate, map[string]interface{}{
		"name":     name,
		"varNames": strings.Join(enumVarNames, ", "),
	})
}
