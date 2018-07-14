package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/count"
	"github.com/jkomoros/boardgame/moves/groups"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/workfit/tester/assert"
	"strconv"
	"testing"
)

//+autoreader
type moveNoOpFixUp struct {
	FixUp
}

func (m *moveNoOpFixUp) Apply(state boardgame.State) error {
	return nil
}

//+autoreader
type moveNoOpFixUpMulti struct {
	FixUpMulti
}

func (m *moveNoOpFixUpMulti) Apply(state boardgame.State) error {
	return nil
}

func TestMoveProgression(t *testing.T) {

	numMoveNames := 3

	singleMoveNames := make([]string, numMoveNames)

	for i := 0; i < numMoveNames; i++ {
		singleMoveNames[i] = strconv.Itoa(i)
	}

	multiMoveNames := make([]string, len(singleMoveNames))

	for i, name := range singleMoveNames {
		multiMoveNames[i] = name + " Multi"
	}

	var configs []GroupableMoveConfig

	for _, name := range singleMoveNames {
		configs = append(configs, newMoveConfig(name, new(moveNoOpFixUp), nil))
	}
	for _, name := range multiMoveNames {
		configs = append(configs, newMoveConfig(name, new(moveNoOpFixUpMulti), nil))
	}

	singleMoveConfigs := make([]interfaces.MoveProgressionGroup, len(singleMoveNames))
	multiMoveConfigs := make([]interfaces.MoveProgressionGroup, len(multiMoveNames))

	for i, _ := range singleMoveNames {
		singleMoveConfigs[i] = configs[i]
		multiMoveConfigs[i] = configs[numMoveNames+i]
	}

	tests := []struct {
		tape           []string
		pattern        []interfaces.MoveProgressionGroup
		expectedResult bool
	}{
		{
			[]string{
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				singleMoveConfigs[0],
				singleMoveConfigs[1],
				singleMoveConfigs[2],
			},
			true,
		},
		{
			[]string{
				singleMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				singleMoveConfigs[0],
				singleMoveConfigs[1],
				singleMoveConfigs[2],
			},
			false,
		},
		{
			[]string{
				singleMoveNames[0],
				singleMoveNames[0],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				singleMoveConfigs[0],
				singleMoveConfigs[1],
				singleMoveConfigs[2],
			},
			false,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				singleMoveConfigs[1],
				singleMoveConfigs[2],
			},
			true,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				multiMoveConfigs[1],
				singleMoveConfigs[2],
			},
			true,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				multiMoveConfigs[1],
				singleMoveConfigs[2],
			},
			false,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[1],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				multiMoveConfigs[1],
				multiMoveConfigs[0],
				multiMoveConfigs[2],
			},
			true,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[1],
				multiMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				multiMoveConfigs[1],
				multiMoveConfigs[0],
				multiMoveConfigs[2],
			},
			true,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				multiMoveNames[1],
				multiMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				multiMoveConfigs[0],
				multiMoveConfigs[1],
				multiMoveConfigs[0],
				multiMoveConfigs[2],
			},
			false,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
				singleMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Serial(
					multiMoveConfigs[0],
					singleMoveConfigs[1],
				),
				singleMoveConfigs[2],
				groups.Serial(
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
			},
			true,
		},
		{
			[]string{
				multiMoveNames[0],
				multiMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Serial(
					multiMoveConfigs[0],
					singleMoveConfigs[1],
				),
				singleMoveConfigs[2],
				groups.Serial(
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
			},
			false,
		},
		{
			//Check in-order OK
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Check partial match is OK
			[]string{
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Check out-of-order OK
			[]string{
				singleMoveNames[1],
				singleMoveNames[2],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Check some multi OK
			[]string{
				multiMoveNames[1],
				multiMoveNames[2],
				multiMoveNames[2],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
					multiMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Check some multi OK but not out of order
			[]string{
				multiMoveNames[1],
				multiMoveNames[2],
				multiMoveNames[2],
				multiMoveNames[0],
				multiMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
					multiMoveConfigs[2],
				),
			},
			false,
		},
		{
			//Check two parallels in a row
			[]string{
				multiMoveNames[1],
				multiMoveNames[2],
				multiMoveNames[0],
				multiMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
					multiMoveConfigs[2],
				),
				groups.Parallel(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
					multiMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Check parallel followed by a serial
			[]string{
				singleMoveNames[1],
				singleMoveNames[2],
				singleMoveNames[0],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
				groups.Serial(
					singleMoveConfigs[2],
					singleMoveConfigs[1],
				),
			},
			true,
		},
		{
			//Parallel that contains a serial
			[]string{
				multiMoveNames[2],
				multiMoveNames[1],
				multiMoveNames[2],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					multiMoveConfigs[0],
					groups.Serial(
						multiMoveConfigs[1],
						multiMoveConfigs[2],
					),
					multiMoveConfigs[2],
				),
			},
			true,
		},
		{
			//Test a parallel with a serial where the beginning of the serial also matches another item.
			[]string{
				multiMoveNames[2],
				multiMoveNames[2],
				multiMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					multiMoveConfigs[2],
					groups.Serial(
						multiMoveConfigs[2],
						multiMoveConfigs[1],
					),
				),
			},
			true,
		},
		{
			//Check parallel with any
			[]string{
				singleMoveNames[1],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.ParallelCount(
					count.Any(),
					singleMoveConfigs[0],
					singleMoveConfigs[1],
				),
				singleMoveConfigs[2],
			},
			true,
		},
		{
			//Ensure that only one can return
			[]string{
				singleMoveNames[1],
				singleMoveNames[0],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.ParallelCount(
					count.Any(),
					singleMoveConfigs[0],
					singleMoveConfigs[1],
				),
				singleMoveConfigs[2],
			},
			false,
		},
	}

	//Note that the old test, progressionMatches() didn't check which types
	//were allowed to have multiple in a row; it assumed that was always OK,
	//and its Legal check of the last item in the containing function made sue
	//that we didn't lay down a move after itself if that wasn't possible; but
	//we assumed that the whole tape up until that point was valid implicitly.
	//But now we check explicitly every time through when one make apply
	//multiple times.

	for i, test := range tests {

		group := groups.Serial(test.pattern...)

		err := matchTape(group, test.tape)

		assert.For(t, i).ThatActual(err == nil).Equals(test.expectedResult)
	}

}
