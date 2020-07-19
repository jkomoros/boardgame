package record

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
)

var fullEncoder = &fullEncoderImpl{}

//This encoder is so easy!
type fullEncoderImpl struct{}

func (f *fullEncoderImpl) CreatePatch(lastState, state boardgame.StateStorageRecord) ([]byte, error) {
	return state, nil
}

func (f *fullEncoderImpl) ConfirmPatch(before, after, formattedPatch []byte) error {

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

func (f *fullEncoderImpl) ApplyPatch(lastStateBlob, patchBlob []byte) (boardgame.StateStorageRecord, error) {
	return patchBlob, nil
}

//Note: this only works if examplePatch is the first one now.
func (f *fullEncoderImpl) Matches(examplePatch []byte) error {

	//We match if the patch has a version string who is an int.

	var probeStruct struct {
		Game []interface{}
	}

	if err := json.Unmarshal(examplePatch, &probeStruct); err == nil {
		return errors.New("Unmarshal probe for Game as multiple items did NOT fail as expected")
	}

	return nil

}
