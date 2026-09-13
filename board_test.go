package boardgame

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
)

func TestBoard(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState := game.CurrentState().ImmutableGameState().(*testGameState)

	board := gameState.MyBoard

	for i, space := range board.Spaces() {
		assert.For(t).ThatActual(space.Board()).Equals(board)
		assert.For(t).ThatActual(space.BoardIndex()).Equals(i)
		assert.For(t).ThatActual(space.Resizable()).IsFalse()
	}

}

func TestEnumBoardUsesValueOrderRatherThanNumericKey(t *testing.T) {
	set := enum.NewSet()
	boardEnum := set.MustAdd("sparse", map[enum.EnumKey]string{2: "Two", 10: "Ten"})
	deck := NewDeck()
	board := deck.NewBoardForEnum(boardEnum, 1)

	if board.SpaceAtKey(2) != board.SpaceAt(0) || board.SpaceAtKey(10) != board.SpaceAt(1) {
		t.Fatal("enum-keyed board did not use enum.Values order")
	}
	if board.SpaceAtKey(3) != nil {
		t.Fatal("enum-keyed board accepted an invalid key")
	}

	blob, err := json.Marshal(board)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(blob), `"Enum":"sparse"`) || !strings.Contains(string(blob), `"Keys":["Two","Ten"]`) {
		t.Fatalf("board JSON did not expose enum identity: %s", blob)
	}
}

func TestBoardRejectsWrongPersistedLengthAndBoundaryIndex(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, _ := concreteStates(game.CurrentState())
	board := gameState.MyBoard
	if board.SpaceAt(board.Len()) != nil || board.ImmutableSpaceAt(board.Len()) != nil {
		t.Fatal("board accepted index equal to Len")
	}

	blob, err := json.Marshal(board)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw boardJSONObj
	if err := json.Unmarshal(blob, &raw); err != nil {
		t.Fatalf("decode wrapper: %v", err)
	}
	raw.Spaces = append(raw.Spaces, raw.Spaces[0])
	corrupt, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal corrupt wrapper: %v", err)
	}
	if err := json.Unmarshal(corrupt, board); err == nil || !strings.Contains(err.Error(), "persisted spaces") {
		t.Fatalf("wrong-length board error = %v", err)
	}
}

func TestEnumBoardRejectsPersistedKeyRemapping(t *testing.T) {
	set := enum.NewSet()
	boardEnum := set.MustAdd("pile", map[enum.EnumKey]string{2: "Left", 10: "Right"})
	original := NewDeck().NewBoardForEnum(boardEnum, 1)
	blob, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	renamed := enum.NewSet().MustAdd("pile", map[enum.EnumKey]string{2: "Right", 10: "Left"})
	destination := NewDeck().NewBoardForEnum(renamed, 1)
	if err := json.Unmarshal(blob, destination); err == nil || !strings.Contains(err.Error(), "key order") {
		t.Fatalf("silently remapped stored piles: %v", err)
	}
	var legacy map[string]json.RawMessage
	if err := json.Unmarshal(blob, &legacy); err != nil {
		t.Fatal(err)
	}
	delete(legacy, "Enum")
	delete(legacy, "Keys")
	blob, err = json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, destination); err != nil {
		t.Fatalf("rejected pre-enum legacy board: %v", err)
	}
}
