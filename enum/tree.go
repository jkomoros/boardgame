package enum

//TreeEnum is a special type of Enum where the list of values also have a tree
//structure that can be interrogated. TreeEnums always have 0 map to "" as the
//root value.
type TreeEnum interface {
	Enum

	//IsLeaf returns true if the given value in the enum represents a leaf (as
	//opposed to branch)
	IsLeaf(val int) bool

	//Children returns all of the val beneath this branch that are direct
	//descendents, either including or excluding non-leaf nodes.
	Children(node int, includeBranches bool) []int

	//Descendants returns all enumvals beneath this point, recursively.
	Descendants(node int, includeBranches bool) []int

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

	//Children is a convenience for val.Enum().TreeEnum().Children(val).
	Children(includeBranches bool) []int

	//Descendants is a convenience for val.Enum().TreeEnum().Descendents(val).
	Descendants(includeBranches bool) []int

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

func (e *enum) IsLeaf(val int) bool {
	if e.tree == nil {
		return false
	}

	return len(e.tree[val]) == 0
}

func (e *enum) Children(node int, includeBranches bool) []int {
	if e.tree == nil {
		return nil
	}

	var result []int

	for _, val := range e.tree[node] {
		if !includeBranches && !e.IsLeaf(val) {
			continue
		}
		result = append(result, val)
	}

	return result
}

func (e *enum) Descendants(node int, includeBranches bool) []int {
	if e.tree == nil {
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
		result = append(result, e.Descendants(val, includeBranches)...)
	}

	return result

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

func (v *variable) Children(includeBranches bool) []int {
	return v.Enum().TreeEnum().Children(v.Value(), includeBranches)
}

func (v *variable) Descendants(includeBranches bool) []int {
	return v.Enum().TreeEnum().Descendants(v.Value(), includeBranches)
}

func (v *variable) NodeString() string {
	//TODO: actually implement this properly
	return v.String()
}
