package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"strings"
	"testing"
)

const (
	colorRed = iota
	colorBlue
	colorGreen
)

var testEnums = enum.NewSet()

var testColorEnum = testEnums.MustAdd("color", map[int]string{
	colorRed:   "Red",
	colorBlue:  "Blue",
	colorGreen: "Green",
})

const (
	phase = iota
	phaseSetUp
	phaseNormal
	phaseNormalPlayerStart
	phaseNormalActivateCard
	phaseScoring
)

var testPhaseEnum = testEnums.MustAddTree("phase",
	map[int]string{
		phase:                   "",
		phaseSetUp:              "Set Up",
		phaseNormal:             "Normal",
		phaseNormalActivateCard: "Activate Card",
		phaseNormalPlayerStart:  "Player Start",
		phaseScoring:            "Scoring",
	},
	map[int]int{
		phase:                   phase,
		phaseSetUp:              phase,
		phaseNormal:             phase,
		phaseNormalPlayerStart:  phaseNormal,
		phaseNormalActivateCard: phaseNormal,
		phaseScoring:            phase,
	})

func defaultTestGameDelegate(extraComponentsToCreate int) *testGameDelegate {

	moveInstaller := func(manager *GameManager) *MoveTypeConfigBundle {

		bundle := NewMoveTypeConfigBundle()

		bundle.AddMoves(
			&testMoveConfig,
			&testMoveIncrementCardInHandConfig,
			&testMoveDrawCardConfig,
			&testMoveAdvanceCurrentPlayerConfig,
			&testMoveInvalidPlayerIndexConfig,
			&testMoveMakeIllegalPhaseConfig,
		)

		return bundle
	}

	return &testGameDelegate{
		moveInstaller:           moveInstaller,
		extraComponentsToCreate: extraComponentsToCreate,
	}
}

func newTestGameManger(t *testing.T) *GameManager {

	manager, err := NewGameManager(defaultTestGameDelegate(0), newTestStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	return manager
}

type nilStackGameDelegate struct {
	testGameDelegate
	//We wil always return a player state that has nil. But if this is false, we will also return one for game.
	nilForPlayer bool
}

func (n *nilStackGameDelegate) PlayerStateConstructor(playe PlayerIndex) ConfigurablePlayerState {
	return &testPlayerState{}
}

func (n *nilStackGameDelegate) GameStateConstructor() ConfigurableSubState {
	if n.nilForPlayer {
		//return a non-nil one.
		return n.testGameDelegate.GameStateConstructor()
	}

	return &testGameState{}
}

type testMoveFailValidConfiguration struct {
	baseFixUpMove
}

func (t *testMoveFailValidConfiguration) ValidConfiguration(exampleState State) error {
	return errors.New("Invalid configuration")
}

func (t *testMoveFailValidConfiguration) Apply(state State) error {
	return nil
}

func (t *testMoveFailValidConfiguration) Legal(state ImmutableState, proposer PlayerIndex) error {
	return nil
}

func (t *testMoveFailValidConfiguration) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testMoveFailValidConfiguration) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMoveFailValidConfiguration) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

var testMoveFailValidConfigurationConfig = MoveTypeConfig{
	Name: "Fail Valid Configuration",
	MoveConstructor: func() Move {
		return new(testMoveFailValidConfiguration)
	},
}

func TestMoveFailsValidConfiguration(t *testing.T) {

	moveInstaller := func(manager *GameManager) *MoveTypeConfigBundle {
		bundle := NewMoveTypeConfigBundle()

		bundle.AddMoves(
			&testMoveFailValidConfigurationConfig,
		)

		return bundle
	}

	_, err := NewGameManager(&testGameDelegate{moveInstaller: moveInstaller}, newTestStorageManager())

	assert.For(t).ThatActual(err).IsNotNil()

}

func TestDefaultMove(t *testing.T) {
	//Tests that Moves based on DefaultMove copy correctly

	game := testGame(t)

	if err := game.SetUp(0, nil, nil); err != nil {
		t.Fatal("Couldn't set up game: " + err.Error())
	}

	//FixUpMoveByName calls Copy under the covers.
	move := game.MoveByName("Advance Current Player")

	assert.For(t).ThatActual(move).IsNotNil()

	assert.For(t).ThatActualString(move.Info().Type().Name()).Equals("Advance Current Player")

	convertedMove, ok := move.(*testMoveAdvanceCurentPlayer)

	assert.For(t).ThatActual(ok).Equals(true)
	assert.For(t).ThatActual(convertedMove).IsNotNil()
}

func TestNilStackErrors(t *testing.T) {

	_, err := NewGameManager(&nilStackGameDelegate{}, newTestStorageManager())

	//We expect to find the error of the nil stack at NewGameManager time,
	//because that's when we validate constructors.
	assert.For(t).ThatActual(err).IsNotNil()

	//playerState will already work for nil stacks, so no need to flip the
	//delegate's behavior to test playerstate like we used to do here.

}

func TestGameManagerModifiableGame(t *testing.T) {
	game := testGame(t)

	game.SetUp(0, nil, nil)

	manager := game.Manager()

	//use ToLower to verify that ID comparison is not case sensitive.
	otherGame := manager.ModifiableGame(strings.ToLower(game.Id()))

	if game != otherGame {
		t.Error("ModifiableGame didn't give back the same game that already existed")
	}

	//OK, forget about the real game to test us making a new one.
	manager.modifiableGames = make(map[string]*Game)

	otherGame = manager.ModifiableGame(game.Id())

	if otherGame == nil {
		t.Error("Other game didn't return anything even though it was in storage!")
	}

	if game == otherGame {
		t.Error("ModifiableGame didn't grab a game from storage fresh")
	}

	otherGame = manager.ModifiableGame("NOGAMEATTHISID")

	if otherGame != nil {
		t.Error("ModifiableGame returned a game even for an invalid ID")
	}

}

func TestGameManagerSetUp(t *testing.T) {

	manager := newTestGameManger(t)

	moves := manager.MoveTypes()

	if moves == nil {
		t.Error("Got nil player moves even after setting up")
	}

	if manager.Agents() == nil {
		t.Error("Agents after setup was nil")
	}

	if manager.AgentByName("test") == nil {
		t.Error("Agent test after setup was nil")
	}

	if manager.MoveTypeByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if manager.MoveTypeByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

}
