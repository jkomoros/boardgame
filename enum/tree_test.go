package enum

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasicTree(t *testing.T) {

	set := NewSet()

	/*

		A (1)
			A (4)
		B (2)
			BA (5)
			BB (6)
				BBA (10)
				BBB (11)
			BC (7)
		C (3)
			CA (8)
			CB (9)

	*/

	values := map[int]string{
		1: "A",
		2: "B",
		3: "C",
		//This is the same value as node 1, to make sure the node values don't
		//have to be unique as long as their fully qualified names are.
		4:  "A",
		5:  "BA",
		6:  "BB",
		7:  "BC",
		8:  "CA",
		9:  "CB",
		10: "BBA",
		11: "BBB",
	}

	parents := map[int]int{
		1:  0,
		2:  0,
		3:  0,
		4:  1,
		5:  2,
		6:  2,
		7:  2,
		8:  3,
		9:  3,
		10: 6,
		11: 6,
	}

	tree, err := set.AddTree("test", values, parents)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(tree).IsNotNil()

	assert.For(t).ThatActual(tree.Parent(0)).Equals(0)
	assert.For(t).ThatActual(tree.String(0)).Equals("")

	assert.For(t).ThatActual(tree.Parent(10)).Equals(6)

	assert.For(t).ThatActual(tree.Ancestors(11)).Equals([]int{0, 2, 6, 11})

	assert.For(t).ThatActual(tree.Children(2, false)).Equals([]int{5, 7})
	assert.For(t).ThatActual(tree.Children(2, true)).Equals([]int{5, 6, 7})
	assert.For(t).ThatActual(tree.Children(0, true)).Equals([]int{1, 2, 3})

	assert.For(t).ThatActual(tree.IsLeaf(10)).IsTrue()
	assert.For(t).ThatActual(tree.IsLeaf(0)).IsFalse()
	assert.For(t).ThatActual(tree.IsLeaf(2)).IsFalse()

	assert.For(t).ThatActual(tree.Descendants(2, false)).Equals([]int{5, 10, 11, 7})
	assert.For(t).ThatActual(tree.Descendants(2, true)).Equals([]int{5, 6, 10, 11, 7})

	assert.For(t).ThatActual(tree.BranchDefaultValue(0)).Equals(4)
	assert.For(t).ThatActual(tree.BranchDefaultValue(2)).Equals(5)
	assert.For(t).ThatActual(tree.BranchDefaultValue(6)).Equals(10)
	assert.For(t).ThatActual(tree.BranchDefaultValue(10)).Equals(10)

	assert.For(t).ThatActual(tree.DefaultValue()).Equals(4)

	assert.For(t).ThatActual(tree.String(2)).Equals("B")
	assert.For(t).ThatActual(tree.String(10)).Equals("B > BB > BBA")
	assert.For(t).ThatActual(tree.ValueFromString("B > BB > BBA")).Equals(10)

	assert.For(t).ThatActual(tree.MustNewTreeVal(10).NodeString()).Equals("BBA")
	assert.For(t).ThatActual(tree.MustNewTreeVal(0).NodeString()).Equals("")
	assert.For(t).ThatActual(tree.MustNewTreeVal(2).NodeString()).Equals("B")

}

func TestBadTreeConfig(t *testing.T) {
	tests := []struct {
		values              map[int]string
		parents             map[int]int
		expectedErrorString string
	}{
		{
			map[int]string{
				0: "not ''",
				1: "foo",
			},
			map[int]int{
				0: 0,
				1: 0,
			},
			"The root node's value must be ''",
		},
		{
			map[int]string{
				0: "",
				1: "foo",
			},
			map[int]int{
				0: 0,
				1: 0,
			},
			"",
		},
		{
			map[int]string{
				0: "",
				1: "foo",
			},
			map[int]int{
				0: 1,
				1: 0,
			},
			"The root node's parent must be itself",
		},
		{
			map[int]string{
				1: "foo",
			},
			map[int]int{
				0: 0,
				1: 0,
			},
			"",
		},
		{
			map[int]string{
				1: "foo",
			},
			map[int]int{
				1: 0,
			},
			"",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
			},
			map[int]int{
				1: 0,
			},
			"Missing parent information for key: 2",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
			},
			map[int]int{
				1: 2,
				2: 0,
			},
			"",
		},
		{
			map[int]string{
				1: "foo",
			},
			map[int]int{
				1: 0,
				2: 1,
			},
			"Parent information provided for 2 but no corresponding value provided.",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
			},
			map[int]int{
				1: 0,
				2: 3,
			},
			"Entry in parent map names a parent that is not in the enum: 3,2",
		},
		{
			map[int]string{
				1: "foo > bar",
			},
			map[int]int{
				1: 0,
			},
			"The node string value for 1 contains the delimiter expression, which is illegal",
		},
		{
			//Check maximally stretched case for cycle test
			map[int]string{
				1: "foo",
				2: "bar",
				3: "baz",
			},
			map[int]int{
				1: 0,
				2: 1,
				3: 2,
			},
			"",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
				3: "baz",
			},
			map[int]int{
				1: 3,
				2: 1,
				3: 2,
			},
			"Detected a cycle in the parent definitions",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
				3: "baz",
				4: "slam",
			},
			map[int]int{
				1: 3,
				2: 1,
				3: 2,
				4: 0,
			},
			"Detected a cycle in the parent definitions",
		},
		{
			map[int]string{
				1: "foo",
				2: "bar",
				3: "baz",
			},
			map[int]int{
				1: 1,
				2: 1,
				3: 2,
			},
			"A non-root node had itself as its own parent: 1",
		},
	}

	for i, test := range tests {

		set := NewSet()

		tree, err := set.AddTree("test", test.values, test.parents)

		if test.expectedErrorString != "" {
			assert.For(t, i).ThatActual(err).IsNotNil()
			assert.For(t, i).ThatActual(tree).IsNil()
			assert.For(t, i).ThatActual(err.Error()).Equals(test.expectedErrorString)
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
			assert.For(t, i).ThatActual(tree).IsNotNil()
		}

	}

}
