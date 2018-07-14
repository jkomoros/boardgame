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

	noOpMoveName := "No Op"

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

	noNopConfig := newMoveConfig(noOpMoveName, new(NoOp), nil)

	configs = append(configs, noNopConfig)

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
			//Check two parallels in a row, where there's a double in between
			[]string{
				singleMoveNames[1],
				singleMoveNames[2],
				singleMoveNames[0],
				singleMoveNames[0],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				),
				groups.Parallel(
					singleMoveConfigs[0],
					singleMoveConfigs[1],
					singleMoveConfigs[2],
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
		{
			//Basic repeat test
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.Exactly(1),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
			},
			true,
		},
		{
			//Too long of a tape to repeat
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.Exactly(1),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
			},
			false,
		},
		{
			//Partial on the second
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.Exactly(3),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
			},
			true,
		},
		{
			//Partial on the second, using AtMost. in a way that is idiomatic
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.AtMost(2),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
			},
			true,
		},
		{
			//Partial on the second, using AtLeast. in a way that is idiomatic
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.AtLeast(2),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
				singleMoveConfigs[2],
			},
			true,
		},
		{
			//Partial on the second, using AtLeast. in a way that is
			//idiomatic, where the second loop doesn't fully complete.
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.AtLeast(2),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
				singleMoveConfigs[2],
			},
			false,
		},
		{
			//Partial on the second, using AtLeast. in a way that is
			//idiomatic, where the repeat only happens once, not twice.
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.AtLeast(2),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
				singleMoveConfigs[2],
			},
			false,
		},
		{
			//Partial on the second, using AtLeast. in a way that is
			//idiomatic, where the repeat happens three times, which is legal.
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[2],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.AtLeast(2),
					groups.Serial(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
				singleMoveConfigs[2],
			},
			true,
		},
		{
			//Two serial groups in a row, in different orders
			[]string{
				singleMoveNames[0],
				singleMoveNames[1],
				singleMoveNames[1],
				singleMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Repeat(
					count.Exactly(2),
					groups.Parallel(
						singleMoveConfigs[0],
						singleMoveConfigs[1],
					),
				),
			},
			true,
		},
		{
			//Two serial groups in a row with two AllowMulti abutting. Doesn't
			//match because the first group consumes both 1's, leaving none
			//for the next to consume.
			[]string{
				multiMoveNames[0],
				multiMoveNames[1],
				multiMoveNames[1],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Serial(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
				),
				groups.Serial(
					multiMoveConfigs[1],
					multiMoveConfigs[0],
				),
			},
			false,
		},
		{
			//Two serial groups in a row with two AllowMulti abutting, but a
			//NoOp as a guard against the first group matching too greedily.
			[]string{
				multiMoveNames[0],
				multiMoveNames[1],
				noOpMoveName,
				multiMoveNames[1],
				multiMoveNames[0],
			},
			[]interfaces.MoveProgressionGroup{
				groups.Serial(
					multiMoveConfigs[0],
					multiMoveConfigs[1],
				),
				groups.Serial(
					noNopConfig,
					multiMoveConfigs[1],
					multiMoveConfigs[0],
				),
			},
			true,
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

		if !assert.For(t, i).ThatActual(err == nil).Equals(test.expectedResult).Passed() {
			if err != nil {
				t.Log(err.Error())
			}
		}

	}

}
