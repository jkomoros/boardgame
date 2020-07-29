package boardgame

import (
	"errors"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/sirupsen/logrus"
)

//defaultGameDelegate is a clone of base.GameDelegate recreated here because
//we can't import base without a cylic dependency.
type defaultGameDelegate struct {
	manager *GameManager
}

//Diagram returns the string "This should be overriden to render a reasonable state here"
func (d *defaultGameDelegate) Diagram(state ImmutableState) string {
	return "This should be overriden to render a reasonable state here"
}

//DisplayName by default just returns the title-case of Name() that is
//returned from the delegate in use.
func (d *defaultGameDelegate) DisplayName() string {
	return strings.Title(d.Manager().Delegate().Name())
}

//Description defaults to "" if not overriden.
func (d *defaultGameDelegate) Description() string {
	return ""
}

//Manager returns the manager object that was provided to SetManager.
func (d *defaultGameDelegate) Manager() *GameManager {
	return d.manager
}

//SetManager keeps a reference to the passed manager, and returns it when
//Manager() is called.
func (d *defaultGameDelegate) SetManager(manager *GameManager) {
	d.manager = manager
}

//DynamicComponentValuesConstructor returns nil, as not all games have
//DynamicComponentValues. Override this if your game does require
//DynamicComponentValues.
func (d *defaultGameDelegate) DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState {
	return nil
}

type isFixUpper interface {
	IsFixUp() bool
}

func isFixUp(move Move) bool {

	fixUpper, ok := move.(isFixUpper)
	if !ok {
		return false
	}

	return fixUpper.IsFixUp()

}

//The Default ProposeFixUpMove runs through all moves in Moves, in order, and
//returns the first one that returns true from IsFixUp and is legal at the
//current state. In many cases, this behavior should be suficient and need not
//be overwritten. Be extra sure that your FixUpMoves have a conservative Legal
//function, otherwise you could get a panic from applying too many FixUp
//moves. Wil emit debug information about why certain fixup moves didn't apply
//if the Manager's log level is Debug or higher.
func (d *defaultGameDelegate) ProposeFixUpMove(state ImmutableState) Move {

	isDebug := d.Manager().Logger().Level >= logrus.DebugLevel

	var logEntry *logrus.Entry

	if isDebug {
		logEntry = d.Manager().Logger().WithFields(logrus.Fields{
			"game":    state.Game().ID(),
			"version": state.Version(),
		})
		logEntry.Debug("***** ProposeFixUpMove called *****")
	}

	for _, move := range state.Game().Moves() {

		var entry *logrus.Entry
		if isDebug {
			entry = logEntry.WithField("movetype", move.Info().Name())
		}

		if !isFixUp(move) {
			//Not a fix up move
			continue
		}

		err := move.Legal(state, AdminPlayerIndex)

		if err == nil {
			if isDebug {
				entry.Debug(move.Info().Name() + " : MATCH")
			}
			//Found it!
			return move
		}
		if isDebug {
			entry.Debug(move.Info().Name() + " : " + err.Error())
		}
	}
	if isDebug {
		logEntry.Debug("NO MATCH")
	}
	//No moves apply now.
	return nil
}

//CurrentPlayerIndex returns gameState.CurrentPlayer, if that is a PlayerIndex
//property. If not, returns ObserverPlayerIndex.≈
func (d *defaultGameDelegate) CurrentPlayerIndex(state ImmutableState) PlayerIndex {
	index, err := state.ImmutableGameState().Reader().PlayerIndexProp("CurrentPlayer")

	if err != nil {
		//Guess that's not where they store CurrentPlayer.
		return ObserverPlayerIndex
	}

	return index
}

//CurrentPhase by default with return the value of gameState.Phase, if it is
//an enum. If it is not, it will return -1 instead, to make it more clear that
//it's an invalid CurrentPhase (phase 0 is often valid).
func (d *defaultGameDelegate) CurrentPhase(state ImmutableState) int {

	phaseEnum, err := state.ImmutableGameState().Reader().ImmutableEnumProp("Phase")

	if err != nil {
		//Guess it wasn't there
		return -1
	}

	return phaseEnum.Value()

}

//PhaseEnum defaults to the enum named "Phase" which is the convention for the
//name of the Phase enum. moves.Default will handle cases where that isn't a
//valid enum gracefully.
func (d *defaultGameDelegate) PhaseEnum() enum.Enum {
	return d.Manager().Chest().Enums().Enum("Phase")
}

func (d *defaultGameDelegate) DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error) {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return nil, errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

//SanitizatinoPolicy uses struct tags to identify the right policy to apply
//(see the package doc on SanitizationPolicy for how to configure those tags).
//It sees which policies apply given the provided group membership, and then
//returns the LEAST restrictive policy that applies. This behavior is almost
//always what you want; it is rare to need to override this method.
func (d *defaultGameDelegate) SanitizationPolicy(prop StatePropertyRef, groupMembership map[int]bool) Policy {

	manager := d.Manager()

	inflater := manager.Internals().StructInflater(prop)

	if inflater == nil {
		return PolicyInvalid
	}

	policyMap := inflater.PropertySanitizationPolicy(prop.PropName)

	var applicablePolicies []int

	for group, isMember := range groupMembership {

		//The only ones that are in the map should be `true` but sanity check
		//just in case.
		if !isMember {
			continue
		}

		//Only if the policy is actually in the map should we use it
		if policy, ok := policyMap[group]; ok {
			applicablePolicies = append(applicablePolicies, int(policy))
		}
	}

	if len(applicablePolicies) == 0 {
		return PolicyVisible
	}

	sort.Ints(applicablePolicies)

	return Policy(applicablePolicies[0])

}

//ComputedGlobalProperties returns nil.
func (d *defaultGameDelegate) ComputedGlobalProperties(state ImmutableState) PropertyCollection {
	return nil
}

//ComputedPlayerProperties returns nil.
func (d *defaultGameDelegate) ComputedPlayerProperties(player ImmutableSubState) PropertyCollection {
	return nil
}

//BeginSetUp does not do anything and returns nil.
func (d *defaultGameDelegate) BeginSetUp(state State, variant Variant) error {
	//Don't need to do anything by default
	return nil
}

//FinishSetUp doesn't do anything and returns nil.
func (d *defaultGameDelegate) FinishSetUp(state State) error {
	//Don't need to do anything by default
	return nil
}

//defaultCheckGameFinishedDelegate can be private because
//DefaultGameFinished implements the methods by default.
type defaultCheckGameFinishedDelegate interface {
	GameEndConditionMet(state ImmutableState) bool
	PlayerScore(pState ImmutableSubState) int
	LowScoreWins() bool
}

//PlayerGameScorer is an optional interface that can be implemented by
//PlayerSubStates. If it is implemented, defaultGameDelegate's default
//PlayerScore() method will return it.
type playerGameScorer interface {
	//Score returns the overall score for the game for the player at this
	//point in time.
	GameScore() int
}

//CheckGameFinished by default checks delegate.GameEndConditionMet(). If true,
//then it fetches delegate.PlayerScore() for each player and returns all
//players who have the highest score as winners. (If delegate.LowScoreWins()
//is true, instead of highest score, it does lowest score.) To use this
//implementation simply implement those methods. This is sufficient for many
//games, but not all, so sometimes needs to be overriden.
func (d *defaultGameDelegate) CheckGameFinished(state ImmutableState) (finished bool, winners []PlayerIndex) {

	if d.Manager() == nil {
		return false, nil
	}

	//Have to reach up to the manager's delegate to get the thing that embeds
	//us. Don't use the comma-ok pattern because we want to panic with
	//descriptive error if not met.
	checkGameFinished := d.Manager().Delegate().(defaultCheckGameFinishedDelegate)

	if !checkGameFinished.GameEndConditionMet(state) {
		return false, nil
	}

	lowScoreWins := checkGameFinished.LowScoreWins()

	//Game is over. What's the most extreme (max or min, depending on
	//LowScoreWins) score?
	extremeScore := 0

	if lowScoreWins {
		extremeScore = math.MaxInt32
	}

	for _, player := range state.ImmutablePlayerStates() {
		score := checkGameFinished.PlayerScore(player)

		if lowScoreWins {
			if score < extremeScore {
				extremeScore = score
			}
		} else {
			if score > extremeScore {
				extremeScore = score
			}
		}
	}

	//Who has the most extreme score score?
	for i, player := range state.ImmutablePlayerStates() {
		score := checkGameFinished.PlayerScore(player)

		if score == extremeScore {
			winners = append(winners, PlayerIndex(i))
		}
	}

	return true, winners

}

//LowScoreWins is used in defaultGameDelegate's CheckGameFinished. If false
//(default) higher scores are better. If true, however, then lower scores win
//(similar to golf), and all of the players with the lowest score win.
func (d *defaultGameDelegate) LowScoreWins() bool {
	return false
}

//GameEndConditionMet is used in the default CheckGameFinished implementation.
//It should return true when the game is over and ready for scoring.
//CheckGameFinished uses this by default; if you override CheckGameFinished
//you don't need to override this. The default implementation of this simply
//returns false.
func (d *defaultGameDelegate) GameEndConditionMet(state ImmutableState) bool {
	return false
}

//PlayerScore is used in the default CheckGameFinished implementation. It
//should return the score for the given player. CheckGameFinished uses this by
//default; if you override CheckGameFinished you don't need to override this.
//The default implementation returns pState.GameScore() (if pState implements
//the PlayerGameScorer interface), or 0 otherwise.
func (d *defaultGameDelegate) PlayerScore(pState ImmutableSubState) int {
	if scorer, ok := pState.(playerGameScorer); ok {
		return scorer.GameScore()
	}
	return 0
}

//DefaultNumPlayers returns 2.
func (d *defaultGameDelegate) DefaultNumPlayers() int {
	return 2
}

//MinNumPlayers returns 1
func (d *defaultGameDelegate) MinNumPlayers() int {
	return 1
}

//MaxNumPlayers returns 16
func (d *defaultGameDelegate) MaxNumPlayers() int {
	return 16
}

//LegalNumPlayers checks that the number of players is between MinNumPlayers
//and MaxNumPlayers, inclusive. You'd only want to override this if some
//player numbers in that range are not legal, for example a game where only
//even numbers of players may play.
func (d *defaultGameDelegate) LegalNumPlayers(numPlayers int) bool {

	min := d.Manager().Delegate().MinNumPlayers()
	max := d.Manager().Delegate().MaxNumPlayers()

	return numPlayers >= min && numPlayers <= max

}

//Variants returns a VariantConfig with no entries.
func (d *defaultGameDelegate) Variants() VariantConfig {
	return VariantConfig{}
}

func (d *defaultGameDelegate) GroupMembership(pState ImmutableSubState) map[int]bool {
	return nil
}

//ConfigureAgents by default returns nil. If you want agents in your game,
//override this.
func (d *defaultGameDelegate) ConfigureAgents() []Agent {
	return nil
}

//ConfigureEnums simply returns nil. In general you want to override this with
//a body of `return Enums`, if you're using `boardgame-util config` to
//generate your enum set.
func (d *defaultGameDelegate) ConfigureEnums() *enum.Set {
	return nil
}

//ConfigureDecks returns a zero-entry map. You want to override this if you
//have any components in your game (which the vast majority of games do)
func (d *defaultGameDelegate) ConfigureDecks() map[string]*Deck {
	return make(map[string]*Deck)
}

//ConfigureConstants returns a zero-entry map. If you have any constants you
//wa8nt to use client-side or in tag-based struct auto-inflaters, you will want
//to override this.
func (d *defaultGameDelegate) ConfigureConstants() PropertyCollection {
	return nil
}

type testGameDelegate struct {
	defaultGameDelegate
	//if this is higher than 0, then will craete this many extra comoponents
	extraComponentsToCreate int
	moveInstaller           func(manager *GameManager) []MoveConfig
}

func (t *testGameDelegate) ConfigureAgents() []Agent {
	return []Agent{
		&testAgent{},
	}
}

func (t *testGameDelegate) ConfigureEnums() *enum.Set {
	return testEnums
}

func (t *testGameDelegate) ConfigureConstants() PropertyCollection {
	return PropertyCollection{
		"ConstantStackSize": 4,
		"MyBool":            false,
	}
}

func (t *testGameDelegate) PhaseEnum() enum.Enum {
	return testPhaseEnum
}

func (t *testGameDelegate) GroupEnum() enum.Enum {
	return nil
}

func (t *testGameDelegate) ConfigureDecks() map[string]*Deck {
	deck := NewDeck()

	deck.AddComponent(&testingComponent{
		String:  "foo",
		Integer: 1,
	})

	deck.AddComponent(&testingComponent{
		String:  "bar",
		Integer: 2,
	})

	deck.AddComponent(&testingComponent{
		String:  "baz",
		Integer: 5,
	})

	deck.AddComponent(&testingComponent{
		String:  "slam",
		Integer: 10,
	})

	deck.AddComponent(&testingComponent{
		String:  "basic",
		Integer: 8,
	})

	for i := 0; i < t.extraComponentsToCreate; i++ {
		deck.AddComponent(&testingComponent{
			String:  "Extra",
			Integer: 8 + i,
		})
	}

	deck.SetGenericValues(&testShadowValues{
		Message: "Foo",
	})

	return map[string]*Deck{
		"test": deck,
	}

}

func (t *testGameDelegate) ConfigureMoves() []MoveConfig {
	return t.moveInstaller(t.Manager())
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error) {
	game, _ := concreteStates(state)
	return game.DrawDeck, nil
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) ComputedGlobalProperties(state ImmutableState) PropertyCollection {
	_, playerStates := concreteStates(state)

	allScores := 0

	for _, player := range playerStates {

		allScores += player.Score
	}

	return PropertyCollection{
		"SumAllScores": allScores,
	}
}

func (t *testGameDelegate) ComputedPlayerProperties(player ImmutableSubState) PropertyCollection {

	playerState := player.(*testPlayerState)

	effectiveMovesLeftThisTurn := playerState.MovesLeftThisTurn

	//Players with Isfoo get a bonus.
	if playerState.IsFoo {
		effectiveMovesLeftThisTurn += 5
	}

	return PropertyCollection{
		"EffectiveMovesLeftThisTurn": effectiveMovesLeftThisTurn,
	}
}

func (t *testGameDelegate) DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState {
	if deck.Name() == "test" {
		return &testingComponentDynamic{
			Stack: deck.NewSizedStack(1),
			Enum:  testColorEnum.NewVal(),
		}
	}
	return nil
}

func (t *testGameDelegate) CheckGameFinished(state ImmutableState) (bool, []PlayerIndex) {
	_, players := concreteStates(state)

	var winners []PlayerIndex

	for i, player := range players {
		if player.Score >= 5 {
			//This user won!
			winners = append(winners, PlayerIndex(i))

			//Keep going through to see if anyone else won at the same time
		}
	}

	if len(winners) > 0 {
		return true, winners
	}

	return false, nil
}

func (t *testGameDelegate) DefaultNumPlayers() int {
	return 3
}

func (t *testGameDelegate) MinNumPlayers() int {
	return 1
}

func (t *testGameDelegate) MaxNumPlayers() int {
	return 5
}

func (t *testGameDelegate) PlayerMayBeActive(player ImmutableSubState) bool {
	return true
}

func (t *testGameDelegate) Variants() VariantConfig {

	return VariantConfig{
		"color": {
			Values: map[string]*VariantDisplayInfo{
				"blue": nil,
				"red": {
					DisplayName: "Red",
					Description: "The color red",
				},
			},
		},
	}
}

func (t *testGameDelegate) BeginSetUp(state State, variant Variant) error {
	game, players := concreteStates(state)

	if len(players) != 3 {
		return errors.New("Only three players are supported")
	}

	game.MyEnumValue.SetValue(colorGreen)

	players[0].MovesLeftThisTurn = 1
	players[2].IsFoo = true
	players[1].EnumVal.SetValue(colorGreen)
	return nil
}

func (t *testGameDelegate) FinishSetUp(state State) error {

	//Set all IntVar's to 1 for dynamic values for all hands. This will help
	//us verify when they are being sanitized.

	game, _ := concreteStates(state)

	for _, c := range game.DrawDeck.Components() {
		values := c.DynamicValues().(*testingComponentDynamic)

		values.IntVar = 1
		values.Enum.SetValue(colorBlue)
	}

	return game.DrawDeck.ComponentAt(game.DrawDeck.Len() - 1).MoveToFirstSlot(game.MyBoard.SpaceAt(1))
}

func (t *testGameDelegate) CurrentPlayerIndex(state ImmutableState) PlayerIndex {
	game, _ := concreteStates(state)

	return game.CurrentPlayer
}

func (t *testGameDelegate) GameStateConstructor() ConfigurableSubState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")
	return &testGameState{
		CurrentPlayer:      0,
		DrawDeck:           deck.NewStack(0),
		Timer:              NewTimer(),
		MyIntSlice:         make([]int, 0),
		MyBoolSlice:        make([]bool, 0),
		MyStringSlice:      make([]string, 0),
		MyPlayerIndexSlice: make([]PlayerIndex, 0),
		MyEnumValue:        testColorEnum.NewVal(),
		MyEnumConst:        testColorEnum.MustNewImmutableVal(colorBlue),
	}
}

func (t *testGameDelegate) PlayerStateConstructor(player PlayerIndex) ConfigurableSubState {
	return &testPlayerState{
		playerIndex: player,
	}
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger(t)

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
