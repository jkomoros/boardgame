package filesystem

import (
	"encoding/json"
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

type record struct {
	Game           *boardgame.GameStorageRecord
	states         []json.RawMessage
	expandedStates []map[string]interface{}
	Moves          []*boardgame.MoveStorageRecord
	StatePatches   []json.RawMessage
}

func (r *record) AddState(state json.RawMessage) error {

	lastState, err := r.State(len(r.StatePatches) - 1)

	if err != nil {
		return errors.New("Couldn't fetch last state: " + err.Error())
	}

	differ := gojsondiff.New()

	patch, err := differ.Compare(lastState, state)

	if err != nil {
		return err
	}

	f := formatter.NewDeltaFormatter()

	js, err := f.FormatAsJson(patch)

	if err != nil {
		return errors.New("Couldn't format patch as json: " + err.Error())
	}

	formattedPatch, err := json.Marshal(js)

	if err != nil {
		return errors.New("Couldn't format patch json to byte: " + err.Error())
	}

	r.states = append(r.states, state)
	r.StatePatches = append(r.StatePatches, formattedPatch)

	return nil

}

//State is a method, not just a normal array, because we actually serialize to
//disk state patches, given that states are big objects and they rarely
//change. So when you request a State we have to derive what it is by applying
//all of the patches up in order.
func (r *record) State(version int) (json.RawMessage, error) {

	if version < 0 {
		//The base object that version 0 is diffed against is the empty object
		return json.RawMessage(`{}`), nil
	}

	if len(r.states) > version {
		return r.states[version], nil
	}

	//Otherwise, derive forward, recursively.

	lastStateBlob, err := r.State(version - 1)

	if err != nil {
		//Don't decorate the error because it will likely stack
		return nil, err
	}

	unmarshaller := gojsondiff.NewUnmarshaller()

	patch, err := unmarshaller.UnmarshalBytes(r.StatePatches[version])

	if err != nil {
		return nil, err
	}

	differ := gojsondiff.New()

	var state map[string]interface{}

	if err := json.Unmarshal(lastStateBlob, &state); err != nil {
		return nil, errors.New("Couldn't unmarshal last blob: " + err.Error())
	}

	differ.ApplyPatch(state, patch)

	blob, err := json.MarshalIndent(state, "", "\t")

	if err != nil {
		return nil, errors.New("Couldn't marshal modified blob: " + err.Error())
	}

	r.states = append(r.states, blob)

	return blob, nil

}
