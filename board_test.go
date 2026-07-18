package boardgame

import (
	"encoding/json"
	"strings"
	"testing"

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
