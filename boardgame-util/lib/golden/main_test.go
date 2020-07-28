package golden

import (
	"encoding/json"
	"flag"
	"testing"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/workfit/tester/assert"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files if they're different instead of erroring")

func TestBasic(t *testing.T) {

	//If we also used updateGolden here, then the two tests would collide.
	err := Compare(blackjack.NewDelegate(), "test/basic_blackjack.json", false)

	assert.For(t).ThatActual(err).IsNil()

}

func TestFolder(t *testing.T) {
	err := CompareFolder(blackjack.NewDelegate(), "test", *updateGolden)
	assert.For(t).ThatActual(err).IsNil()
}

func TestMoveAlignment(t *testing.T) {
	tests := []struct {
		description string
		new         []*boardgame.MoveStorageRecord
		old         []*boardgame.MoveStorageRecord
		golden      []*boardgame.MoveStorageRecord
	}{
		{
			"Single move no op",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Removed Single move",
			[]*boardgame.MoveStorageRecord{},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{},
		},
		{
			"No move with single added move",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Two move no op",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Two move simple timestamp copy",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Two move splice at front",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(0, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(2, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Two move splice at end",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(2, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(5, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		{
			"Two move singular splice in middle",
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(1, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(5, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(2, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(5, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
			[]*boardgame.MoveStorageRecord{
				{
					Name:      "A",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(3, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "F",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(4, 0),
					Blob:      json.RawMessage("{}"),
				},
				{
					Name:      "B",
					Version:   -1,
					Initiator: 0,
					Phase:     0,
					Proposer:  -2,
					Timestamp: time.Unix(5, 0),
					Blob:      json.RawMessage("{}"),
				},
			},
		},
		//TODO: test a double splice in middle
		//TODO: a test where the Blobs don't align
	}

	for i, test := range tests {
		alignMoveTimes(test.new, test.old)
		assert.For(t, i, test.description).ThatActual(test.new).Equals(test.golden).ThenDiffOnFail()
	}
}
