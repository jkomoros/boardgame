package api

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/werewolf"
)

func werewolfProjectionFixture(t *testing.T, phase string, wolf boardgame.PlayerIndex) (*boardgame.Game, boardgame.State) {
	t.Helper()
	manager, err := boardgame.NewGameManager(werewolf.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	copied, err := game.CurrentState().Copy(false)
	if err != nil {
		t.Fatal(err)
	}
	state, ok := copied.(boardgame.State)
	if !ok {
		t.Fatal("copied Werewolf state is not mutable")
	}
	phaseValue, err := state.GameState().ReadSetter().EnumProp("Phase")
	if err != nil {
		t.Fatal(err)
	}
	if err := phaseValue.SetValue(phaseValue.Enum().ValueFromString(phase)); err != nil {
		t.Fatal(err)
	}
	for i, player := range state.PlayerStates() {
		if err := player.ReadSetter().SetBoolProp("SeatFilled", true); err != nil {
			t.Fatal(err)
		}
		if err := player.ReadSetter().SetBoolProp("SeatClosed", true); err != nil {
			t.Fatal(err)
		}
		role, err := player.ReadSetter().EnumProp("Role")
		if err != nil {
			t.Fatal(err)
		}
		value := role.Enum().ValueFromString("Villager")
		if boardgame.PlayerIndex(i) == wolf {
			value = role.Enum().ValueFromString("Werewolf")
		}
		if err := role.SetValue(value); err != nil {
			t.Fatal(err)
		}
		if err := player.ReadSetter().SetPlayerIndexProp("DayVote", -1); err != nil {
			t.Fatal(err)
		}
		if err := player.ReadSetter().SetPlayerIndexProp("NightVote", -1); err != nil {
			t.Fatal(err)
		}
	}
	return game, state
}

func werewolfChoiceSchema(t *testing.T, game *boardgame.Game, name string) boardgame.MoveChoiceProjectionSchema {
	t.Helper()
	schema, err := boardgame.BuildMoveChoiceProjectionSchema(game.Manager())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range schema {
		if item.MoveName == name {
			if item.FieldName != "VoteTarget" || item.Source != boardgame.MoveChoiceSourcePlayers {
				t.Fatalf("%s schema = %#v", name, item)
			}
			return item
		}
	}
	t.Fatalf("choice schema omitted %q: %#v", name, schema)
	return boardgame.MoveChoiceProjectionSchema{}
}

func candidateAvailability(set *projectedMoveChoiceSet) []bool {
	if set == nil {
		return nil
	}
	result := make([]bool, len(set.Candidates))
	for i, candidate := range set.Candidates {
		result[i] = candidate.Available
	}
	return result
}

func candidatePlayerIndexes(set *projectedMoveChoiceSet) []int {
	if set == nil {
		return nil
	}
	result := make([]int, len(set.Candidates))
	for i, candidate := range set.Candidates {
		value, ok := candidate.Value.(boardgame.PlayerIndex)
		if !ok {
			panic("player candidate had unexpected type")
		}
		result[i] = int(value)
	}
	return result
}

func TestWerewolfVoteChoicesUseCanonicalAnyPlayerLegality(t *testing.T) {
	dayGame, dayState := werewolfProjectionFixture(t, "Day", 0)
	daySchema := werewolfChoiceSchema(t, dayGame, "Cast Vote")
	day, err := projectMoveChoiceSet(dayGame, dayState, 1, daySchema, new(projectedMoveChoiceBudget))
	if err != nil {
		t.Fatal(err)
	}
	wantDay := []bool{true, false, true, true, true}
	if got := candidateAvailability(day); !reflect.DeepEqual(got, wantDay) {
		t.Fatalf("day vote availability = %v, want %v", got, wantDay)
	}

	nightGame, nightState := werewolfProjectionFixture(t, "Night", 0)
	if err := nightState.PlayerStates()[3].ReadSetter().SetBoolProp("Eliminated", true); err != nil {
		t.Fatal(err)
	}
	if err := nightState.PlayerStates()[4].ReadSetter().SetBoolProp("PlayerInactive", true); err != nil {
		t.Fatal(err)
	}
	nightSchema := werewolfChoiceSchema(t, nightGame, "Cast Night Vote")
	night, err := projectMoveChoiceSet(nightGame, nightState, 0, nightSchema, new(projectedMoveChoiceBudget))
	if err != nil {
		t.Fatal(err)
	}
	// Inactive players are omitted from the player-source universe; the
	// eliminated player remains present but canonical Legal disables it.
	wantNight := []bool{false, true, true, false}
	if got := candidateAvailability(night); !reflect.DeepEqual(got, wantNight) {
		t.Fatalf("night wolf vote availability = %v, want %v", got, wantNight)
	}
	if got, want := candidatePlayerIndexes(night), []int{0, 1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("night wolf vote candidates = %v, want sparse roster %v", got, want)
	}

	villager, err := projectMoveChoiceSet(nightGame, nightState, 1, nightSchema, new(projectedMoveChoiceBudget))
	if err != nil {
		t.Fatal(err)
	}
	if villager != nil {
		t.Fatalf("villager received night choices: %#v", villager)
	}
}

func TestWerewolfNightChoicesIgnoreUnrelatedHiddenRoles(t *testing.T) {
	game, state := werewolfProjectionFixture(t, "Night", 0)
	schema := werewolfChoiceSchema(t, game, "Cast Night Vote")
	before, err := projectMoveChoiceSet(game, state, 0, schema, new(projectedMoveChoiceBudget))
	if err != nil {
		t.Fatal(err)
	}

	copied, err := state.Copy(false)
	if err != nil {
		t.Fatal(err)
	}
	hiddenVariant := copied.(boardgame.State)
	role, err := hiddenVariant.PlayerStates()[2].ReadSetter().EnumProp("Role")
	if err != nil {
		t.Fatal(err)
	}
	if err := role.SetValue(role.Enum().ValueFromString("Werewolf")); err != nil {
		t.Fatal(err)
	}
	after, err := projectMoveChoiceSet(game, hiddenVariant, 0, schema, new(projectedMoveChoiceBudget))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := candidateAvailability(after), candidateAvailability(before); !reflect.DeepEqual(got, want) {
		t.Fatalf("night choices changed with unrelated hidden role: got %v, want %v", got, want)
	}
}
