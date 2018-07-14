package moves

import (
	"github.com/jkomoros/boardgame"
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
				groups.Serial{
					multiMoveConfigs[0],
					singleMoveConfigs[1],
				},
				singleMoveConfigs[2],
				groups.Serial{
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				},
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
				groups.Serial{
					multiMoveConfigs[0],
					singleMoveConfigs[1],
				},
				singleMoveConfigs[2],
				groups.Serial{
					singleMoveConfigs[1],
					singleMoveConfigs[2],
				},
			},
			false,
		},
	}

	//TODO: as new group types are added in groups, bring them in for testing here.

	//Note that the old test, progressionMatches() didn't check which types
	//were allowed to have multiple in a row; it assumed that was always OK,
	//and its Legal check of the last item in the containing function made sue
	//that we didn't lay down a move after itself if that wasn't possible; but
	//we assumed that the whole tape up until that point was valid implicitly.
	//But now we check explicitly every time through when one make apply
	//multiple times.

	for i, test := range tests {

		group := groups.Serial(test.pattern)

		err := matchTape(group, test.tape)

		assert.For(t, i).ThatActual(err == nil).Equals(test.expectedResult)
	}

}
