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

//ImmutableTreeVal is a value from a TreeEnum.
type ImmutableTreeVal interface {
	ImmutableVal

	//IsLeaf is a convenience for
	//!val.TreeEnum().IsLeaf(val.Value())
	IsLeaf() bool

	//NodeString returns the name of this specific node, whereas String
	//returns the fully qualified name. So whereas String() might return
	//"Normal - Default - Save Item", NodeString() will return "Default".
	NodeString() string

	//TreeEnum returns the tree enum we're part of. A convenience for
	//val.Enum().TreeEnum().
	TreeEnum() TreeEnum
}

//TreeVal is a value from a tree enum.
type TreeVal interface {
	Val

	//IsLeaf is a convenience for
	//!val.TreeEnum().IsLeaf(val.Value())
	IsLeaf() bool

	//NodeString returns the name of this specific node, whereas String
	//returns the fully qualified name. So whereas String() might return
	//"Normal - Default - Save Item", NodeString() will return "Default".
	NodeString() string

	//TreeEnum returns the tree enum we're part of. A convenience for
	//val.Enum().TreeEnum().
	TreeEnum() TreeEnum
}
