/*

Package test is a package that is used to run a boardgame/server.StorageManager
implementation through its paces and verify it does everything correctly.

*/
package test

import (
	"encoding/json"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/jkomoros/boardgame/server/api"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/workfit/tester/assert"
)

//StorageManager is the interface for everything that needs to be connected.
//It's just api.StorageManager with one more method.
type StorageManager interface {
	api.StorageManager

	//CleanUp will be called when a given manager is done and can be dispoed of.
	CleanUp()
}

type managerMap map[string]*boardgame.GameManager

func (m managerMap) Get(name string) *boardgame.GameManager {
	return m[name]
}

//StorageManagerFactory vends a new StorageManager to test. Users of this
//package provide one.
type StorageManagerFactory func() StorageManager

//Test is the primary entrypoint for this package, running BasicTest, UsersTest,
//AgentsTest, and ListingTest.
func Test(factory StorageManagerFactory, testName string, connectConfig string, t *testing.T) {

	BasicTest(factory, testName, connectConfig, t)
	UsersTest(factory, testName, connectConfig, t)
	AgentsTest(factory, testName, connectConfig, t)
	ListingTest(factory, testName, connectConfig, t)

}

//BasicTest does the basic tests
func BasicTest(factory StorageManagerFactory, testName string, connectConfig string, t *testing.T) {
	storage := factory()

	defer storage.Close()

	defer storage.CleanUp()

	if err := storage.Connect(connectConfig); err != nil {
		t.Fatal("Unexpected error connecting: ", err.Error())
	}

	assert.For(t).ThatActual(storage.Name()).Equals(testName)

	managers := make(managerMap)

	tictactoeManager, _ := boardgame.NewGameManager(tictactoe.NewDelegate(), storage)

	managers[tictactoeManager.Delegate().Name()] = tictactoeManager

	memoryManager, _ := boardgame.NewGameManager(memory.NewDelegate(), storage)

	managers[memoryManager.Delegate().Name()] = memoryManager

	tictactoeGame, err := tictactoeManager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	eGame, err := storage.ExtendedGame(tictactoeGame.ID())

	if err != nil {
		t.Error("Error getting eGame: " + err.Error())
	}

	assert.For(t).ThatActual(eGame).IsNotNil()

	assert.For(t).ThatActual(tictactoeGame.Created().UnixNano()-tictactoeGame.Modified().UnixNano() < 100).IsTrue()

	assert.For(t).ThatActual(eGame.Owner).Equals("")

	eGame.Owner = "Foo"

	lastSeenTimestamp := tictactoeGame.Modified()

	err = storage.UpdateExtendedGame(tictactoeGame.ID(), eGame)

	assert.For(t).ThatActual(err).IsNil()

	newEGame, err := storage.ExtendedGame(tictactoeGame.ID())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(newEGame).Equals(eGame)

	move := tictactoeGame.MoveByName("Place Token")

	if move == nil {
		t.Fatal(testName, "Couldn't find a move")
	}

	if err := <-tictactoeGame.ProposeMove(move, boardgame.AdminPlayerIndex); err != nil {
		t.Fatal(testName, "Couldn't make move", err)
	}

	eGame, err = storage.ExtendedGame(tictactoeGame.ID())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(tictactoeGame.Modified().After(lastSeenTimestamp)).IsTrue()

	refriedMove, err := tictactoeGame.Move(1)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(refriedMove.Info().Version()).Equals(1)

	assert.For(t).ThatActual(refriedMove.TopLevelStruct()).Equals(refriedMove.TopLevelStruct())

	//RefriedMove and Move will have a different TopLevelStruct which is fine,
	//so set them both to nil before comparing.

	refriedMove.SetTopLevelStruct(nil)
	move.SetTopLevelStruct(nil)

	move.SetInfo(nil)
	refriedMove.SetInfo(nil)

	assert.For(t).ThatActual(refriedMove).Equals(move)

	refriedMoves, err := storage.Moves(tictactoeGame.ID(), 0, 2)

	assert.For(t).ThatActual(err).IsNil()

	//The 0'th move doesn't exist
	assert.For(t).ThatActual(len(refriedMoves)).Equals(2)

	//Make sure they're ascending
	lastVersion := 0
	for i, move := range refriedMoves {
		assert.For(t, i).ThatActual(move.Version > lastVersion).IsTrue()
		lastVersion = move.Version
		if i > 0 {
			assert.For(t).ThatActual(move.Phase).Equals(1)
		}
	}

	//OK, now test that the manager and SetUp and everyone did the right thing.

	localGame, err := storage.Game(tictactoeGame.ID())

	if err != nil {
		t.Error(testName, "Unexpected error", err)
	}

	if localGame == nil {
		t.Fatal(testName, "Couldn't get game copy out")
	}

	assert.For(t).ThatActual(tictactoeGame.StorageRecord().SecretSalt).Equals(localGame.SecretSalt)

	blob, err := json.MarshalIndent(tictactoeGame.StorageRecord(), "", "  ")

	if err != nil {
		t.Fatal(testName, "couldn't marshal game", err)
	}

	localBlob, err := json.MarshalIndent(localGame, "", "  ")

	if err != nil {
		t.Fatal(testName, "Couldn't marshal localGame", err)
	}

	compareJSONObjects(blob, localBlob, testName+"Comparing game and local game", t)

	//Verify that if the game is stored with wrong name that doesn't match manager it won't load up.

	variant := boardgame.Variant{
		"numcards": "small",
		"cardset":  "all",
	}

	memoryGame, err := memoryManager.NewGame(0, variant, nil)

	assert.For(t).ThatActual(variant).Equals(memoryGame.Variant())

	refriedMemoryGame := memoryManager.Game(memoryGame.ID())

	assert.For(t).ThatActual(refriedMemoryGame).IsNotNil()

	assert.For(t).ThatActual(refriedMemoryGame.Variant()).Equals(memoryGame.Variant())

	games := storage.ListGames(10, listing.All, "", "")

	if games == nil {
		t.Error(testName, "ListGames gave back nothing")
	}

	if len(games) != 2 {
		t.Error(testName, "We called listgames with a tictactoe game and a blackjack game, but got", len(games), "back.")
	}

	lastSeenTimestamp = time.Date(3000, 1, 1, 1, 1, 1, 1, time.UTC)

	for _, game := range games {
		if game.Modified.After(lastSeenTimestamp) {
			t.Error("Games were not sorted descending by lastSeenTimeStamp")
		}
		lastSeenTimestamp = game.Modified
	}

	//TODO: figure out how to test that name is matched when retrieving from store.

}

//UsersTest does the basic tests of Users
func UsersTest(factory StorageManagerFactory, testName string, connectConfig string, t *testing.T) {
	storage := factory()

	defer storage.Close()

	defer storage.CleanUp()

	if err := storage.Connect(connectConfig); err != nil {
		t.Fatal("Err connecting to storage: ", err)
	}

	manager, _ := boardgame.NewGameManager(tictactoe.NewDelegate(), storage)

	game, _ := manager.NewGame(2, nil, nil)

	var nilIds []string

	ids := storage.UserIDsForGame("DEADBEEF")

	assert.For(t).ThatActual(ids).Equals(nilIds)

	ids = storage.UserIDsForGame(game.ID())

	assert.For(t).ThatActual(ids).Equals([]string{"", ""})

	userID := "THISISAVERYLONGUSERIDTOTESTTHATWEDONTCLIPSHORTUSERIDSTOOAGGRESSIVELY"

	cookie := "MYCOOKIE"

	fetchedUser := storage.GetUserByID(userID)

	var nilUser *users.StorageRecord

	assert.For(t).ThatActual(fetchedUser).Equals(nilUser)

	user := &users.StorageRecord{ID: userID}

	err := storage.UpdateUser(user)

	assert.For(t).ThatActual(err).IsNil()

	fetchedUser = storage.GetUserByID(userID)

	assert.For(t).ThatActual(fetchedUser).Equals(user)

	fetchedUser = storage.GetUserByCookie(cookie)

	assert.For(t).ThatActual(fetchedUser).Equals(nilUser)

	err = storage.ConnectCookieToUser(cookie, user)

	assert.For(t).ThatActual(err).IsNil()

	fetchedUser = storage.GetUserByCookie(cookie)

	assert.For(t).ThatActual(fetchedUser).Equals(user)

	err = storage.SetPlayerForGame(game.ID(), 0, userID)

	assert.For(t).ThatActual(err).IsNil()

	ids = storage.UserIDsForGame(game.ID())

	assert.For(t).ThatActual(ids).Equals([]string{userID, ""})

	err = storage.SetPlayerForGame(game.ID(), 0, userID)

	assert.For(t).ThatActual(err).IsNotNil()
}

//AgentsTest does the basic tests of Agents.
func AgentsTest(factory StorageManagerFactory, testName string, connectConfig string, t *testing.T) {

	storage := factory()

	defer storage.Close()
	defer storage.CleanUp()

	if err := storage.Connect(connectConfig); err != nil {
		t.Fatal("Err connecting to storage: ", err)
	}

	manager, _ := boardgame.NewGameManager(tictactoe.NewDelegate(), storage)

	game, err := manager.NewGame(2, nil, []string{"", "ai"})

	assert.For(t).ThatActual(err).IsNil()

	refriedGame := manager.Game(game.ID())

	assert.For(t).ThatActual(refriedGame.Agents()).Equals(game.Agents())

	refriedBlob, err := storage.AgentState(game.ID(), 0)

	assert.For(t).ThatActual(err).IsNil()

	var nilBlob []byte

	assert.For(t).ThatActual(refriedBlob).Equals(nilBlob)

	blob := []byte("ThisIsABlob")

	err = storage.SaveAgentState(game.ID(), 0, blob)

	assert.For(t).ThatActual(err).IsNil()

	refriedBlob, err = storage.AgentState(game.ID(), 0)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(refriedBlob).Equals(blob)

	newBlob := []byte("ThisIsANewBlob")

	err = storage.SaveAgentState(game.ID(), 0, newBlob)

	assert.For(t).ThatActual(err).IsNil()

	refriedBlob, err = storage.AgentState(game.ID(), 0)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(refriedBlob).Equals(newBlob)

}

//ListingTest does the basic tests of Listing.
func ListingTest(factory StorageManagerFactory, testName string, connectConfig string, t *testing.T) {

	storage := factory()

	defer storage.Close()
	defer storage.CleanUp()

	if err := storage.Connect(connectConfig); err != nil {
		t.Fatal("Err connecting to storage: ", err)
	}

	testUser := "Foo"
	testUserOther := "Bam"
	testUserAnother := "Slam"

	storage.UpdateUser(&users.StorageRecord{
		ID: testUser,
	})

	storage.UpdateUser(&users.StorageRecord{
		ID: testUserOther,
	})

	storage.UpdateUser(&users.StorageRecord{
		ID: testUserAnother,
	})

	manager, _ := boardgame.NewGameManager(tictactoe.NewDelegate(), storage)

	agents := manager.Agents()

	agentName := ""

	if len(agents) > 0 {
		agentName = agents[0].Name()
	}

	blackjackManager, _ := boardgame.NewGameManager(blackjack.NewDelegate(), storage)

	configs := []struct {
		IsBlackjack bool
		UserZero    string
		UserOne     string
		IsAgent     []bool
		Finished    bool
		Open        bool
		Visible     bool
	}{
		//0
		//All: Yes
		//ParticipatingActive: Yes
		//ParticipatingFinished: No, not Finished
		//VisibleJoinableActive: No, testuser is player
		//VisibleActive: No, testUser is player
		//VisibleJoinableActive (No User): Yes
		//VisibleActive (No User): No, open slots
		{
			false,
			testUser,
			"",
			[]bool{false, false},
			false,
			true,
			true,
		},
		//1
		//All: Yes
		//ParticipatingActive: Yes
		//ParticipatingFinished: No, not Finished
		//VisibleJoinableActive: No, testuser is player
		//VisibleActive: No, testUser is player
		//VisibleJoinableActive (No User): Yes
		//VisibleActive (No User): No, open slots
		{
			true,
			testUser,
			"",
			[]bool{false, false},
			false,
			true,
			true,
		},
		//2
		//All: Yes
		//ParticipatingActive: No, not user
		//ParticipatingFinished: No, not user
		//VisibleJoinableActive: Yes
		//VisibleActive: No, game is joinable
		//VisibleJoinableActive (No User): Yes
		//VisibleActive (No User): No, open slots
		{
			false,
			"",
			"",
			[]bool{false, false},
			false,
			true,
			true,
		},
		//3
		//All: Yes
		//ParticipatingActive: Yes
		//ParticipatingFinished: No, not Finished
		//VisibleJoinableActive: No, testuser is player
		//VisibleActive: No, testUser is player
		//VisibleJoinableActive (No User): Yes
		//VisibleActive (No User): No, open slots
		{
			false,
			"",
			testUser,
			[]bool{false, false},
			false,
			true,
			true,
		},
		//4
		//All: Yes
		//ParticipatingActive: No, game finished
		//ParticipatingFinished: Yes
		//VisibleJoinableActive: No, testuser is player
		//VisibleActive: No, testuser is player
		//VisibleJoinableActive (No User): No, Finished
		//VisibleActive (No User): No, Finished
		{
			false,
			testUserOther,
			testUser,
			[]bool{false, false},
			true,
			false,
			false,
		},
		//5
		//All: Yes
		//ParticipatingActive: No, not player
		//ParticipatingFinished: No, not player
		//VisibleJoinableActive: No, game is not visible
		//VisibleActive: No, game is not visible
		//VisibleJoinableActive (No User): No, not visible
		//VisibleActive (No User): No, not visible
		{
			false,
			testUserOther,
			"",
			[]bool{false, false},
			false,
			false,
			false,
		},
		//6
		//All: Yes
		//ParticipatingActive: No, not player
		//ParticipatingFinished: No, not player
		//VisibleJoinableActive: No, no open slots
		//VisibleActive: Yes
		//VisibleJoinableActive (No User): No, no slots
		//VisibleActive (No User): Yes
		{
			false,
			testUserOther,
			testUserAnother,
			[]bool{false, false},
			false,
			true,
			true,
		},
		//7
		//All: Yes
		//ParticipatingActive: No, not player
		//ParticipatingFinished: No, not player
		//VisibleJoinableActive: No, not open
		//VisibleActive: Yes
		//VisibleJoinableActive (No User): No, not open
		//VisibleActive (No User): Yes
		{
			false,
			testUserOther,
			"",
			[]bool{false, false},
			false,
			false,
			true,
		},
		//8
		//All: Yes
		//ParticipatingActive: No, not player
		//ParticipatingFinished: No, not player
		//VisibleJoinableActive: No no slots (one player, one agent)
		//VisibleActive: Yes
		//VisibleJoinableActive (No User): No, no slots
		//VisibleActive (No User): Yes
		{
			false,
			testUserOther,
			"",
			[]bool{false, true},
			false,
			true,
			true,
		},
	}

	goldenRecords := make([]*extendedgame.CombinedStorageRecord, len(configs))

	for i, config := range configs {

		var game *boardgame.Game
		var err error

		var agents []string
		for _, isAgent := range config.IsAgent {
			name := ""
			if isAgent {
				name = agentName
			}
			agents = append(agents, name)
		}

		if config.IsBlackjack {
			game, err = blackjackManager.NewGame(2, nil, agents)
		} else {
			game, err = manager.NewGame(2, nil, agents)
		}

		if err != nil {
			t.Fatal("Couldn't create game: " + err.Error())
		}
		if config.UserZero != "" {
			if err := storage.SetPlayerForGame(game.ID(), 0, config.UserZero); err != nil {
				t.Error("Couldn't store user: " + err.Error())
			}
		}
		if config.UserOne != "" {
			if err := storage.SetPlayerForGame(game.ID(), 1, config.UserOne); err != nil {
				t.Error("Couldn't store user: " + err.Error())
			}
		}
		eGame, err := storage.ExtendedGame(game.ID())
		if err != nil {
			t.Fatal("Couldn't get extended game info: " + err.Error())
		}
		eGame.Open = config.Open
		eGame.Visible = config.Visible
		storage.UpdateExtendedGame(game.ID(), eGame)
		if config.Finished {
			gameRec, err := storage.Game(game.ID())
			if err != nil {
				t.Fatal("Couldn't get gameass: " + err.Error())
			}
			gameRec.Finished = true
			gameRec.Version++
			err = storage.SaveGameAndCurrentState(gameRec, game.CurrentState().StorageRecord(), nil)
			if err != nil {
				t.Fatal("Couldn't save the game: " + err.Error())
			}
		}

		goldenRecords[i], err = storage.CombinedGame(game.ID())

		if err != nil {
			t.Fatal("Couldn't get combined record: ", err.Error())
		}
	}

	expectations := []struct {
		list     listing.Type
		user     string
		gameType string
		result   []int
	}{
		{
			listing.All,
			"",
			"",
			[]int{0, 1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			listing.ParticipatingActive,
			testUser,
			"",
			[]int{0, 1, 3},
		},
		{
			listing.ParticipatingActive,
			"",
			"",
			[]int{},
		},
		{
			listing.ParticipatingFinished,
			testUser,
			"",
			[]int{4},
		},
		{
			listing.ParticipatingFinished,
			"",
			"",
			[]int{},
		},
		{
			listing.VisibleJoinableActive,
			testUser,
			"",
			[]int{2},
		},
		{
			listing.VisibleJoinableActive,
			"",
			"",
			[]int{0, 1, 2, 3},
		},
		{
			listing.VisibleActive,
			testUser,
			"",
			[]int{6, 7, 8},
		},
		{
			listing.VisibleActive,
			"",
			"",
			[]int{6, 7, 8},
		},
		{
			listing.ParticipatingActive,
			testUser,
			"tictactoe",
			[]int{0, 3},
		},
	}

	for i, expectation := range expectations {
		games := storage.ListGames(10, expectation.list, expectation.user, expectation.gameType)

		golden := make([]*extendedgame.CombinedStorageRecord, len(expectation.result))

		for i, index := range expectation.result {
			golden[i] = goldenRecords[index]
		}

		sort.Slice(games, func(i, j int) bool {
			return games[i].ID < games[j].ID
		})
		sort.Slice(golden, func(i, j int) bool {
			return golden[i].ID < golden[j].ID
		})

		if len(games) == 0 && len(golden) == 0 {
			//For some reason assert thinks that two that are both len(0) are not the same, even when they are.
			continue
		}
		if !assert.For(t, i).ThatActual(games).Equals(golden).Passed() {
			//The Diff for when they aren't the same isn't very good, so do our own.
			log.Println("Games")
			for i, item := range games {
				log.Println(i, item, storage.UserIDsForGame(item.ID))
			}
			log.Println("Golden")
			for i, item := range golden {
				log.Println(i, item, storage.UserIDsForGame(item.ID))
			}

		}
	}

}

func compareJSONObjects(in []byte, golden []byte, message string, t *testing.T) {

	//recreated in boardgame/state_test.go

	var deserializedIn interface{}
	var deserializedGolden interface{}

	json.Unmarshal(in, &deserializedIn)
	json.Unmarshal(golden, &deserializedGolden)

	if deserializedIn == nil {
		t.Error("In didn't deserialize", message)
	}

	if deserializedGolden == nil {
		t.Error("Golden didn't deserialize", message)
	}

	if !reflect.DeepEqual(deserializedIn, deserializedGolden) {
		t.Error("Got wrong json.", message, "Got", string(in), "wanted", string(golden))
	}
}
