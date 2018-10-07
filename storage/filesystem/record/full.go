package record

import (
	"encoding/json"
	"errors"
	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	"strings"
)

//This encoder is so easy!
type fullEncoder struct{}

func (f *fullEncoder) CreatePatch(lastState, state boardgame.StateStorageRecord) ([]byte, error) {
	return state, nil
}

func (f *fullEncoder) ConfirmPatch(before, after, formattedPatch []byte) error {

	var inflatedAfter map[string]interface{}
	if err := json.Unmarshal(after, &inflatedAfter); err != nil {
		return errors.New("Couldn't unmarshal before blob: " + err.Error())
	}

	var inflatedPatch map[string]interface{}
	if err := json.Unmarshal(formattedPatch, &inflatedPatch); err != nil {
		return errors.New("Couldn't unmarshal patch blob: " + err.Error())
	}

	if diff := deep.Equal(inflatedAfter, inflatedPatch); len(diff) > 0 {

		return errors.New("Patched before did not equal after: " + strings.Join(diff, "\n"))

	}

	return nil
}

func (f *fullEncoder) ApplyPatch(lastStateBlob, patchBlob []byte) (boardgame.StateStorageRecord, error) {
	return patchBlob, nil
}
