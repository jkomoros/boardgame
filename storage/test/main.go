/*

	test is a package that is used to run a boardgame/server.StorageManager
	implementation through its paces and verify it does everything correctly.

*/
package test

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/workfit/tester/assert"
	"reflect"
	"testing"
)

type StorageManager interface {
	boardgame.StorageManager

	//CleanUp will be called when a given manager is done and can be dispoed of.
	CleanUp()

	//The methods past this point are the same ones that are included in Server.StorageManager
	Name() string

	Close()

	ListGames(max int) []*boardgame.GameStorageRecord

	UserIdsForGame(gameId string) []string

	SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error

	UpdateUser(user *users.StorageRecord) error

	GetUserById(uid string) *users.StorageRecord

	GetUserByCookie(cookie string) *users.StorageRecord

	ConnectCookieToUser(cookie string, user *users.StorageRecord) error
}

type managerMap map[string]*boardgame.GameManager

func (m managerMap) Get(name string) *boardgame.GameManager {
	return m[name]
}

type StorageManagerFactory func() StorageManager

func Test(factory StorageManagerFactory, testName string, t *testing.T) {

	BasicTest(factory, testName, t)
	UsersTest(factory, testName, t)

}

func BasicTest(factory StorageManagerFactory, testName string, t *testing.T) {
	storage := factory()

	assert.For(t).ThatActual(storage.Name()).DoesNotEqual("")

	managers := make(managerMap)

	tictactoeManager := tictactoe.NewManager(storage)

	managers[tictactoeManager.Delegate().Name()] = tictactoeManager

	blackjackManager := blackjack.NewManager(storage)

	managers[blackjackManager.Delegate().Name()] = blackjackManager

	tictactoeGame := boardgame.NewGame(tictactoeManager)

	tictactoeGame.SetUp(0)

	move := tictactoeGame.PlayerMoveByName("Place Token")

	if move == nil {
		t.Fatal(testName, "Couldn't find a move")
	}

	if err := <-tictactoeGame.ProposeMove(move, boardgame.AdminPlayerIndex); err != nil {
		t.Fatal(testName, "Couldn't make move", err)
	}

	//OK, now test that the manager and SetUp and everyone did the right thing.

	localGame, err := storage.Game(tictactoeGame.Id())

	if err != nil {
		t.Error(testName, "Unexpected error", err)
	}

	if localGame == nil {
		t.Fatal(testName, "Couldn't get game copy out")
	}

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

	blackjackGame := boardgame.NewGame(blackjackManager)

	blackjackGame.SetUp(0)

	games := storage.ListGames(10)

	if games == nil {
		t.Error(testName, "ListGames gave back nothing")
	}

	if len(games) != 2 {
		t.Error(testName, "We called listgames with a tictactoe game and a blackjack game, but got", len(games), "back.")
	}

	//TODO: figure out how to test that name is matched when retrieving from store.

	storage.CleanUp()
}

func UsersTest(factory StorageManagerFactory, testName string, t *testing.T) {
	storage := factory()

	manager := tictactoe.NewManager(storage)

	game := boardgame.NewGame(manager)

	game.SetUp(2)

	var nilIds []string

	ids := storage.UserIdsForGame("DEADBEEF")

	assert.For(t).ThatActual(ids).Equals(nilIds)

	ids = storage.UserIdsForGame(game.Id())

	assert.For(t).ThatActual(ids).Equals([]string{"", ""})

	userId := "MYUSERID"

	cookie := "MYCOOKIE"

	fetchedUser := storage.GetUserById(userId)

	var nilUser *users.StorageRecord

	assert.For(t).ThatActual(fetchedUser).Equals(nilUser)

	user := &users.StorageRecord{Id: userId}

	err := storage.UpdateUser(user)

	assert.For(t).ThatActual(err).IsNil()

	fetchedUser = storage.GetUserById(userId)

	assert.For(t).ThatActual(fetchedUser).Equals(user)

	fetchedUser = storage.GetUserByCookie(cookie)

	assert.For(t).ThatActual(fetchedUser).Equals(nilUser)

	err = storage.ConnectCookieToUser(cookie, user)

	assert.For(t).ThatActual(err).IsNil()

	fetchedUser = storage.GetUserByCookie(cookie)

	assert.For(t).ThatActual(fetchedUser).Equals(user)

	err = storage.SetPlayerForGame(game.Id(), 0, userId)

	assert.For(t).ThatActual(err).IsNil()

	ids = storage.UserIdsForGame(game.Id())

	assert.For(t).ThatActual(ids).Equals([]string{userId, ""})

	err = storage.SetPlayerForGame(game.Id(), 0, userId)

	assert.For(t).ThatActual(err).IsNotNil()
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
