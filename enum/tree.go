package enum

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

//The string used to join the individual node names into one (e.g. "Normal - Deal Cards - Deal To First Player")
const treeValStringJoiner = " - "

//TreeEnum is a special type of Enum where the list of values also have a tree
//structure that can be interrogated. TreeEnums always have 0 map to "" as the
//root value.
type TreeEnum interface {
	Enum

	//IsLeaf returns true if the given value in the enum represents a leaf (as
	//opposed to branch)
	IsLeaf(val int) bool

	//Parent returns the value of the node who is the direct parent. The root
	//will return itself.
	Parent(val int) int

	//Ancestors returns the path from the root down to the given value,
	//INCLUDING the value itself.
	Ancestors(val int) []int

	//Children returns all of the val beneath this branch that are direct
	//descendents, either including or excluding non-leaf nodes.
	Children(node int, includeBranches bool) []int

	//Descendants returns all enumvals beneath this point, recursively.
	Descendants(node int, includeBranches bool) []int

	//BranchDefaultVal is like DefaultVal, but only for nodes underneath this
	//node. It returns the left-most leaf node under this node. If node is a
	//leaf, it returns itself. Otherwise, it returns the BranchDefaultVal of
	//its first child.
	BranchDefaultVal(node int) int

	NewImmutableTreeVal(val int) (ImmutableTreeVal, error)
	NewTreeVal() TreeVal

	MustNewImmutableTreeVal(val int) ImmutableTreeVal
	MustNewTreeVal(val int) TreeVal
}

//TreeValGetters is the collection of methods that TreeVals have beyodn normal
//Vals. It is factored out into a separate interface to clarify how
//ImmutableTreeVal and TreeVal differ from their non-treeval types.
type TreeValGetters interface {
	//IsLeaf is a convenience for val.Enum().TreeEnum().IsLeaf(val.Value())
	IsLeaf() bool

	//Parent is a convenience for val.Enum().TreeEnum().Parent(val).
	Parent() int

	//Ancestors is a convenience for val.Enum().TreeEnum().Ancestors(val).
	Ancestors() []int

	//Children is a convenience for val.Enum().TreeEnum().Children(val).
	Children(includeBranches bool) []int

	//Descendants is a convenience for val.Enum().TreeEnum().Descendents(val).
	Descendants(includeBranches bool) []int

	//BranchDefaultVal is a convenience for
	//val.Enum().TreeEnum().BranchDefaulVal(val).
	BranchDefaultVal() int

	//NodeString returns the name of this specific node, whereas String
	//returns the fully qualified name. So whereas String() might return
	//"Normal - Default - Save Item", NodeString() will return "Default".
	NodeString() string
}

//ImmutableTreeVal is a value from a TreeEnum.
type ImmutableTreeVal interface {
	ImmutableVal
	TreeValGetters
}

//TreeVal is a value from a tree enum.
type TreeVal interface {
	Val
	TreeValGetters
}

//MustAddTree is like AddTree, but instead of an error it will panic if the
//enum cannot be added. This is useful for defining your enums at the package
//level outside of an init().
func (e *Set) MustAddTree(enumName string, values map[int]string, parents map[int]int) TreeEnum {
	result, err := e.AddTree(enumName, values, parents)

	if err != nil {
		panic("Couldn't add to enumset: " + err.Error())
	}

	return result
}

//AddTree adds a new tree enum. Unlike a normal enum, you also must pass amap
//of value to its parent. Every value in values must also be present in
//parents. In a TreeEnum, the root is always value 0 and always has string
//value "". You may omit the root value if you choose because it is implied.
//The string value map should be just the name of the node itself. The
//effective name of the node will be joined all of its ancestors. Typically
//you rely on autoreader to create these on your behalf, because the initial
//set-up is finicky with the two maps.
func (s *Set) AddTree(enumName string, values map[int]string, parents map[int]int) (TreeEnum, error) {

	str, ok := values[0]

	if ok {
		if str != "" {
			return nil, errors.New("The root node's value must be ''")
		}
	} else {
		values[0] = ""
	}

	parent, ok := parents[0]

	if ok {
		if parent != 0 {
			return nil, errors.New("The root node's parent must be itself")
		}
	} else {
		parents[0] = 0
	}

	//Verify that values and parents inclue the same keys--that is, each item denotes its parent.
	for val, _ := range values {
		if _, ok := parents[val]; !ok {
			return nil, errors.New("Missing parent information for key: " + strconv.Itoa(val))
		}
	}
	for parent, _ := range parents {
		if _, ok := values[parent]; !ok {
			return nil, errors.New("Parent information provided for " + strconv.Itoa(parent) + " but no corresponding value provided.")
		}
	}

	//Verify that each parent corresponds to a value in the values map.
	for child, parent := range parents {
		if _, ok := values[parent]; !ok {
			return nil, errors.New("Entry in parent map names a parent that is not in the enum: " + strconv.Itoa(parent) + "," + strconv.Itoa(child))
		}
	}

	//Verify that the string values don't contain the delimiter sequence
	for val, str := range values {
		if strings.Contains(str, treeValStringJoiner) {
			return nil, errors.New("The node string value for " + strconv.Itoa(val) + " contains the delimiter expression, which is illegal")
		}
	}

	//Check for cycles in graph

	//We know the root connects to itself since we already verfied that. So
	//keep track of each number and whether it connects to root.
	connectedToRoot := make(map[int]bool, len(parents))

	for child, parent := range parents {
		if connectedToRoot[child] {
			//If we already saw that we were connected to root while proving
			//others were OK, we're done.
			continue
		}
		if connectedToRoot[parent] {
			//If we already know our parent is connected to root, then we're
			//done--but take note first that we are also connected to root.
			connectedToRoot[child] = true
			continue
		}
		//OK, haven't seen either, we need to walk up until we get to root.
		visitedNodes := make(map[int]bool, len(parents))
		currentNode := child
		for currentNode != 0 {
			visitedNodes[currentNode] = true
			currentNode = parents[currentNode]
			if visitedNodes[currentNode] {
				return nil, errors.New("Detected a cycle in the parent definitions")
			}
		}
		//CurrentNode is 0, which means all visited nodes are connectedToRoot.
		for visitedNode, _ := range visitedNodes {
			connectedToRoot[visitedNode] = true
		}
	}

	//Preprocess to create the tree map
	childrenMap := make(map[int][]int, len(parents))
	for child, parent := range parents {
		if child == 0 {
			continue
		}
		childrenMap[parent] = append(childrenMap[parent], child)
	}
	for node, _ := range childrenMap {
		//TODO: verify that this actually sorts in place
		sort.Ints(childrenMap[node])
	}

	e, err := s.addEnumImpl(enumName, values)

	if err != nil {
		return nil, err
	}

	e.children = childrenMap
	e.parents = parents

	return e, nil

}

func (e *enum) IsLeaf(val int) bool {
	if e.children == nil {
		return false
	}

	return len(e.children[val]) == 0
}

func (e *enum) Parent(val int) int {
	return e.parents[val]
}

func (e *enum) Ancestors(val int) []int {
	//Base case
	if val == 0 {
		return []int{0}
	}

	return append(e.Ancestors(e.Parent(val)), val)

}

func (e *enum) Children(node int, includeBranches bool) []int {
	if e.children == nil {
		return nil
	}

	var result []int

	for _, val := range e.children[node] {
		if !includeBranches && !e.IsLeaf(val) {
			continue
		}
		result = append(result, val)
	}

	return result
}

func (e *enum) Descendants(node int, includeBranches bool) []int {
	if e.children == nil {
		return nil
	}
	if e.IsLeaf(node) {
		return []int{}
	}

	var result []int

	for _, child := range e.Children(node, true) {
		result = append(result, e.descendantsRecursive(child, includeBranches)...)
	}

	return result
}

func (e *enum) descendantsRecursive(node int, includeBranches bool) []int {
	if e.children == nil {
		return nil
	}

	if e.IsLeaf(node) {
		return []int{node}
	}

	var result []int

	if includeBranches {
		result = []int{node}
	}

	for _, val := range e.Children(node, true) {
		result = append(result, e.descendantsRecursive(val, includeBranches)...)
	}

	return result

}

func (e *enum) BranchDefaultVal(node int) int {
	if e.children == nil {
		return e.DefaultValue()
	}

	if e.IsLeaf(node) {
		return node
	}

	return e.BranchDefaultVal(e.Children(node, true)[0])
}

func (e *enum) NewImmutableTreeVal(val int) (ImmutableTreeVal, error) {
	v := e.NewTreeVal()
	if err := v.SetValue(val); err != nil {
		return nil, err
	}
	return v, nil
}

func (e *enum) NewTreeVal() TreeVal {
	return &variable{
		e,
		e.DefaultValue(),
	}
}

func (e *enum) MustNewImmutableTreeVal(val int) ImmutableTreeVal {
	return e.MustNewTreeVal(val)
}

func (e *enum) MustNewTreeVal(val int) TreeVal {
	v := e.NewTreeVal()
	if err := v.SetValue(val); err != nil {
		panic("Couldn't set string value: " + err.Error())
	}
	return v
}

func (v *variable) IsLeaf() bool {
	return v.Enum().TreeEnum().IsLeaf(v.Value())
}

func (v *variable) Parent() int {
	return v.Enum().TreeEnum().Parent(v.Value())
}

func (v *variable) Ancestors() []int {
	return v.Enum().TreeEnum().Ancestors(v.Value())
}

func (v *variable) Children(includeBranches bool) []int {
	return v.Enum().TreeEnum().Children(v.Value(), includeBranches)
}

func (v *variable) Descendants(includeBranches bool) []int {
	return v.Enum().TreeEnum().Descendants(v.Value(), includeBranches)
}

func (v *variable) BranchDefaultVal() int {
	return v.Enum().TreeEnum().BranchDefaultVal(v.Value())
}

func (v *variable) NodeString() string {
	//TODO: actually implement this properly
	return v.String()
}
