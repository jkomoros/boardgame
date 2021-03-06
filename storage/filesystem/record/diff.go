package record

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var diffEncoder = &diffEncoderImpl{}

type diffEncoderImpl struct{}

func (d *diffEncoderImpl) CreatePatch(lastState, state boardgame.StateStorageRecord) ([]byte, error) {
	differ := gojsondiff.New()

	patch, err := differ.Compare(lastState, state)

	if err != nil {
		return nil, err
	}

	f := formatter.NewDeltaFormatter()

	js, err := f.FormatAsJson(patch)

	if err != nil {
		return nil, errors.New("Couldn't format patch as json: " + err.Error())
	}

	formattedPatch, err := json.Marshal(js)

	if err != nil {
		return nil, errors.New("Couldn't format patch json to byte: " + err.Error())
	}

	return formattedPatch, nil
}

func (d *diffEncoderImpl) ConfirmPatch(before, after, formattedPatch []byte) error {
	var inflatedBefore map[string]interface{}
	if err := json.Unmarshal(before, &inflatedBefore); err != nil {
		return errors.New("Couldn't unmarshal before blob: " + err.Error())
	}

	var inflatedAfter map[string]interface{}
	if err := json.Unmarshal(after, &inflatedAfter); err != nil {
		return errors.New("Couldn't unmarshal before blob: " + err.Error())
	}

	unmarshaller := gojsondiff.NewUnmarshaller()

	reinflatedPatch, err := unmarshaller.UnmarshalBytes(formattedPatch)
	if err != nil {
		return errors.New("Couldn't reinflate patch: " + err.Error())
	}

	differ := gojsondiff.New()
	differ.ApplyPatch(inflatedBefore, reinflatedPatch)

	if diff := deep.Equal(inflatedBefore, inflatedAfter); len(diff) > 0 {

		return errors.New("Patched before did not equal after: " + strings.Join(diff, "\n"))

	}

	return nil
}

func (d *diffEncoderImpl) ApplyPatch(lastStateBlob, patchBlob []byte) (boardgame.StateStorageRecord, error) {
	unmarshaller := gojsondiff.NewUnmarshaller()

	patch, err := unmarshaller.UnmarshalBytes(patchBlob)

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

	return blob, nil
}

//Note: this only works if examplePatch is the first one now.
func (d *diffEncoderImpl) Matches(examplePatch []byte) error {

	//We match if the patch has a version string who is an int.

	var probeStruct struct {
		Game []interface{}
	}

	if err := json.Unmarshal(examplePatch, &probeStruct); err != nil {
		return errors.New("Unmarshal probe for Game as a list failed: " + err.Error())
	}

	return nil

}
