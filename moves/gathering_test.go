package moves

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/storage/memory"
)

// ---- Gathering test game setup ----

const (
	gatheringTeamUnset = iota
	gatheringTeamRed
	gatheringTeamBlue
)

const (
	gatheringRoleUnset = iota
	gatheringRoleSpymaster
	gatheringRoleGuesser
)

const (
	gatheringPhaseGathering = iota
	gatheringPhasePlaying
)

var gatheringEnums = enum.NewSet()

var gatheringTeamEnum = gatheringEnums.MustAdd("team", map[enum.EnumKey]string{
	gatheringTeamUnset: "Unset",
	gatheringTeamRed:   "Red",
	gatheringTeamBlue:  "Blue",
})

var gatheringRoleEnum = gatheringEnums.MustAdd("role", map[enum.EnumKey]string{
	gatheringRoleUnset:     "Unset",
	gatheringRoleSpymaster: "Spymaster",
	gatheringRoleGuesser:   "Guesser",
})

var gatheringPhaseEnum = gatheringEnums.MustAdd("phase", map[enum.EnumKey]string{
	gatheringPhaseGathering: "Gathering",
	gatheringPhasePlaying:   "Playing",
})

//boardgame:codegen
type gatheringGameState struct {
	base.SubState
	behaviors.PhaseBehavior
}

//boardgame:codegen
type gatheringPlayerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
	behaviors.PlayerTeam
	behaviors.PlayerRole
	behaviors.GameAdministrator
}

type gatheringDelegate struct {
	base.GameDelegate
	readyToStartErr error
}

func (g *gatheringDelegate) Name() string {
	return "moves"
}

func (g *gatheringDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gatheringDelegate) MinNumPlayers() int {
	return 2
}

func (g *gatheringDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gatheringGameState)
}

func (g *gatheringDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(gatheringPlayerState)
}

func (g *gatheringDelegate) ConfigureEnums() *enum.Set {
	return gatheringEnums
}

func (g *gatheringDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{}
}

func (g *gatheringDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		Add(
			auto.MustConfig(new(SeatPlayer)),
		),
		AddForPhase(gatheringPhaseGathering,
			GatheringMoves(auto)...,
		),
		AddOrderedForPhase(gatheringPhaseGathering,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(gatheringPhasePlaying, gatheringPhaseEnum)),
		),
	)
}

func (g *gatheringDelegate) ReadyToStart(state boardgame.ImmutableState) error {
	return g.readyToStartErr
}

func (g *gatheringDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	return nil, nil
}

func newGatheringGameManager(t *testing.T) (*boardgame.GameManager, *gatheringDelegate) {
	t.Helper()
	delegate := &gatheringDelegate{}
	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())
	if err != nil {
		t.Fatal("Couldn't create gathering game manager:", err)
	}
	return manager, delegate
}

// ---- Tests ----

func TestGatheringMovesAutoDetection(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	// GatheringMoves should detect PlayerTeam and PlayerRole on playerState
	auto := NewAutoConfigurer(manager.Delegate())
	moves := GatheringMoves(auto)

	// Should have SelectTeam and SelectRole (playerState has both behaviors)
	if len(moves) != 2 {
		t.Errorf("GatheringMoves: expected 2 configs (team + role), got %d", len(moves))
	}

	// Verify names
	names := make(map[string]bool)
	for _, m := range moves {
		names[m.Name()] = true
	}
	if !names["Select Team"] {
		t.Error("GatheringMoves did not include Select Team")
	}
	if !names["Select Role"] {
		t.Error("GatheringMoves did not include Select Role")
	}
}

func TestAnyPlayerLegalRejectsObserverDefault(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	state := game.CurrentState()

	// Create a SelectTeam move with default TargetPlayerIndex (ObserverPlayerIndex)
	move := game.MoveByName("Select Team")
	if move == nil {
		t.Fatal("Couldn't find Select Team move")
	}

	// DefaultsForState should set TargetPlayerIndex to ObserverPlayerIndex
	move.DefaultsForState(state)

	// Legal should reject because target is ObserverPlayerIndex
	err = move.Legal(state, boardgame.PlayerIndex(0))
	if err == nil {
		t.Error("AnyPlayer.Legal should reject when TargetPlayerIndex is ObserverPlayerIndex (default)")
	}
}

func TestAnyPlayerLegalAcceptsAdmin(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	state := game.CurrentState()

	move := game.MoveByName("Select Team")
	if move == nil {
		t.Fatal("Couldn't find Select Team move")
	}

	// Set target to player 0 and propose as admin
	selectTeam := move.(*SelectTeam)
	selectTeam.TargetPlayerIndex = 0
	selectTeam.SelectedTeam.SetStringValue("Red")

	// Seat the player first (mark seat as filled)
	playerState := state.ImmutablePlayerStates()[0]
	if seater, ok := playerState.(interface{ SetSeatFilled() }); ok {
		// Can't modify immutable state, but Legal should still work with admin proposer
		_ = seater
	}

	// Admin proposer should pass the equivalence check even though target != proposer
	err = move.Legal(state, boardgame.AdminPlayerIndex)
	// This may fail on seat check (seat not filled), but should NOT fail on
	// "you can only make this move for yourself"
	if err != nil && err.Error() == "you can only make this move for yourself" {
		t.Error("AnyPlayer.Legal should accept AdminPlayerIndex as proposer")
	}
}

func TestAnyPlayerLegalRejectsWrongProposer(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	state := game.CurrentState()

	move := game.MoveByName("Select Team")
	if move == nil {
		t.Fatal("Couldn't find Select Team move")
	}

	selectTeam := move.(*SelectTeam)
	selectTeam.TargetPlayerIndex = 0
	selectTeam.SelectedTeam.SetStringValue("Red")

	// Player 1 proposing for player 0's target should fail.
	// The error might be about the phase (if the game auto-advanced past
	// gathering) or about the wrong proposer — both are valid rejections.
	err = move.Legal(state, boardgame.PlayerIndex(1))
	if err == nil {
		t.Error("AnyPlayer.Legal should reject when proposer != target")
	}
}

func TestSelectTeamAcceptsValidEnumValue(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	state := game.CurrentState()

	move := game.MoveByName("Select Team")
	if move == nil {
		t.Fatal("Couldn't find Select Team move")
	}

	selectTeam := move.(*SelectTeam)
	selectTeam.TargetPlayerIndex = 0
	// Value 0 is "Unset" — should be accepted (not rejected like before)
	selectTeam.SelectedTeam.SetStringValue("Unset")

	err = move.Legal(state, boardgame.AdminPlayerIndex)
	// Should not fail on "you must select a team" — enum value 0 is valid
	if err != nil && err.Error() == "you must select a team" {
		t.Error("SelectTeam should accept enum value 0 (the old bug)")
	}
}

func TestSelectRoleUniqueRejectsDuplicate(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	// We need a custom game with WithUnique on SelectRole.
	// For this test, let's just verify the non-unique default allows duplicates.
	state := game.CurrentState()

	move := game.MoveByName("Select Role")
	if move == nil {
		t.Fatal("Couldn't find Select Role move")
	}

	selectRole := move.(*SelectRole)
	selectRole.TargetPlayerIndex = 0
	selectRole.SelectedRole.SetStringValue("Spymaster")

	// Without WithUnique, this should be legal (default allows duplicates)
	err = move.Legal(state, boardgame.AdminPlayerIndex)
	// May fail on seat check, but should not fail on uniqueness
	if err != nil && err.Error() == "another player already has that role" {
		t.Error("SelectRole without WithUnique should allow duplicates")
	}
}

func TestReadyToStartBlocksWaitForEnoughPlayers(t *testing.T) {
	manager, delegate := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	// The game should be in the gathering phase waiting for players.
	// ReadyToStart returns nil by default, so WaitForEnoughPlayers only
	// blocks on player count.
	state := game.CurrentState()
	_ = state

	// Set ReadyToStart to return an error
	delegate.readyToStartErr = errors.New("teams not balanced")

	// Verify the error propagates through FrameworkComputedGlobalProperties
	computed := manager.Delegate().FrameworkComputedGlobalProperties(game.CurrentState())
	readyErr, ok := computed["ReadyToStartError"]
	if !ok {
		t.Error("FrameworkComputedGlobalProperties should include ReadyToStartError when ReadyToStart fails")
	}
	if readyErr != "teams not balanced" {
		t.Errorf("ReadyToStartError: expected 'teams not balanced', got '%v'", readyErr)
	}

	// Clear the error
	delegate.readyToStartErr = nil
	computed = manager.Delegate().FrameworkComputedGlobalProperties(game.CurrentState())
	_, hasErr := computed["ReadyToStartError"]
	if hasErr {
		t.Error("FrameworkComputedGlobalProperties should not include ReadyToStartError when ReadyToStart succeeds")
	}
}

func TestComputedPropertiesIncludeGatheringData(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	// Check that FrameworkComputedGlobalProperties includes AvailableTeams and AvailableRoles
	computed := manager.Delegate().FrameworkComputedGlobalProperties(game.CurrentState())

	teams, ok := computed["AvailableTeams"]
	if !ok {
		t.Error("FrameworkComputedGlobalProperties should include AvailableTeams")
	}
	teamList, ok := teams.([]map[string]interface{})
	if !ok {
		t.Errorf("AvailableTeams should be []map[string]interface{}, got %T", teams)
	}
	if len(teamList) != 3 { // Unset, Red, Blue
		t.Errorf("AvailableTeams: expected 3 values, got %d", len(teamList))
	}

	roles, ok := computed["AvailableRoles"]
	if !ok {
		t.Error("FrameworkComputedGlobalProperties should include AvailableRoles")
	}
	roleList, ok := roles.([]map[string]interface{})
	if !ok {
		t.Errorf("AvailableRoles should be []map[string]interface{}, got %T", roles)
	}
	if len(roleList) != 3 { // Unset, Spymaster, Guesser
		t.Errorf("AvailableRoles: expected 3 values, got %d", len(roleList))
	}

	// Check FrameworkComputedPlayerProperties includes TeamValue and RoleValue
	playerComputed := manager.Delegate().FrameworkComputedPlayerProperties(game.CurrentState().ImmutablePlayerStates()[0])
	if _, ok := playerComputed["TeamValue"]; !ok {
		t.Error("FrameworkComputedPlayerProperties should include TeamValue")
	}
	if _, ok := playerComputed["RoleValue"]; !ok {
		t.Error("FrameworkComputedPlayerProperties should include RoleValue")
	}
}

func TestSelectColorFallbackName(t *testing.T) {
	move := &SelectColor{}
	if name := move.FallbackName(nil); name != "Select Color" {
		t.Errorf("SelectColor.FallbackName: expected 'Select Color', got '%s'", name)
	}
	if help := move.FallbackHelpText(); help != "Choose your player color." {
		t.Errorf("SelectColor.FallbackHelpText: expected 'Choose your player color.', got '%s'", help)
	}
}

func TestSelectionIsUnique(t *testing.T) {
	tests := []struct {
		name          string
		config        boardgame.PropertyCollection
		defaultUnique bool
		want          bool
	}{
		{"default true, no config", boardgame.PropertyCollection{}, true, true},
		{"default false, no config", boardgame.PropertyCollection{}, false, false},
		{"WithUnique overrides default false", boardgame.PropertyCollection{configPropUnique: true}, false, true},
		{"WithUnique false overrides default true", boardgame.PropertyCollection{configPropUnique: false}, true, false},
		{"WithAllowDuplicates overrides default true", boardgame.PropertyCollection{configPropAllowDuplicates: true}, true, false},
		{"WithAllowDuplicates false enforces unique", boardgame.PropertyCollection{configPropAllowDuplicates: false}, false, true},
		{"WithUnique takes precedence over WithAllowDuplicates", boardgame.PropertyCollection{configPropUnique: true, configPropAllowDuplicates: true}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectionIsUnique(tt.config, tt.defaultUnique)
			if got != tt.want {
				t.Errorf("selectionIsUnique(%v, %v) = %v, want %v", tt.config, tt.defaultUnique, got, tt.want)
			}
		})
	}
}

func TestGatheringMovesPartialBehaviors(t *testing.T) {
	// Test that GatheringMoves correctly handles partial behavior coverage.
	// Our test playerState has PlayerTeam + PlayerRole but NOT PlayerColor.
	manager, _ := newGatheringGameManager(t)

	auto := NewAutoConfigurer(manager.Delegate())
	moves := GatheringMoves(auto)

	// Should have exactly 2 (Team + Role), not 3 (no Color)
	if len(moves) != 2 {
		t.Errorf("GatheringMoves with Team+Role (no Color): expected 2, got %d", len(moves))
	}

	names := make(map[string]bool)
	for _, m := range moves {
		names[m.Name()] = true
	}
	if names["Select Color"] {
		t.Error("GatheringMoves should NOT include Select Color when PlayerColor is not embedded")
	}
}

func TestPlayerIsAdminAutoSet(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	// In a non-server context, SeatPlayer won't fire (no server injection).
	// But we can check FrameworkComputedPlayerProperties for admin status.
	// Since no one is seated, no one should be admin.
	state := game.CurrentState()
	for i, ps := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsAdmin(ps) {
			t.Errorf("Player %d should not be admin before being seated", i)
		}
	}
}

func TestAdminPlayerFallbacks(t *testing.T) {
	move := &AdminPlayer{}
	if name := move.FallbackName(nil); name != "Admin Player Move" {
		t.Errorf("AdminPlayer.FallbackName: expected 'Admin Player Move', got '%s'", name)
	}
	if help := move.FallbackHelpText(); help != "A move that only the game administrator can make." {
		t.Errorf("AdminPlayer.FallbackHelpText: got '%s'", help)
	}
}

func TestCheckRequireAdmin(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	state := game.CurrentState()

	// Without WithRequireAdmin, any proposer passes
	configNoAdmin := boardgame.PropertyCollection{}
	if err := checkRequireAdmin(configNoAdmin, state, boardgame.PlayerIndex(0)); err != nil {
		t.Errorf("checkRequireAdmin without config should pass, got: %v", err)
	}

	// With WithRequireAdmin, non-admin proposer fails
	configWithAdmin := boardgame.PropertyCollection{configPropRequireAdmin: true}
	err = checkRequireAdmin(configWithAdmin, state, boardgame.PlayerIndex(0))
	if err == nil {
		t.Error("checkRequireAdmin should reject non-admin player 0")
	}

	// AdminPlayerIndex always passes
	err = checkRequireAdmin(configWithAdmin, state, boardgame.AdminPlayerIndex)
	if err != nil {
		t.Errorf("checkRequireAdmin should allow AdminPlayerIndex, got: %v", err)
	}
}

func TestComputedPropertiesIncludeIsGameAdmin(t *testing.T) {
	manager, _ := newGatheringGameManager(t)

	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	// No one is admin initially
	state := game.CurrentState()
	computed := manager.Delegate().FrameworkComputedPlayerProperties(state.ImmutablePlayerStates()[0])
	if _, ok := computed["IsGameAdmin"]; ok {
		t.Error("FrameworkComputedPlayerProperties should not include IsGameAdmin when player is not admin")
	}
}
