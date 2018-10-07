package record

import (
	"encoding/json"
	"errors"
	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	jd "github.com/josephburnett/jd/lib"
	"strings"
)

type josephBurnettEncoder struct{}

func jdDiffToJson(renderedDiff string) ([]byte, error) {
	rows := strings.Split(renderedDiff, "\n")

	return json.MarshalIndent(rows, "", "\t")
}

func jsonBlobToDiff(blob []byte) (jd.Diff, error) {
	var rows []string

	if err := json.Unmarshal(blob, &rows); err != nil {
		return nil, errors.New("Couldn't unmarshal jd patch: " + err.Error())
	}

	diff, err := jd.ReadDiffString(strings.Join(rows, "\n"))

	if err != nil {
		return nil, errors.New("Couldn't parse jd diff from string: " + err.Error())
	}
	return diff, nil
}

func (j *josephBurnettEncoder) CreatePatch(lastState, state boardgame.StateStorageRecord) ([]byte, error) {
	lastStateNode, err := jd.ReadJsonString(string(lastState))
	if err != nil {
		return nil, errors.New("Couldn't prase lastState: " + err.Error())
	}

	stateNode, err := jd.ReadJsonString(string(state))

	if err != nil {
		return nil, errors.New("Couldn't parse state: " + err.Error())
	}

	diff := lastStateNode.Diff(stateNode)

	return jdDiffToJson(diff.Render())

}

func (j *josephBurnettEncoder) ConfirmPatch(before, after, formattedPatch []byte) error {

	inflatedBefore, err := jd.ReadJsonString(string(before))
	if err != nil {
		return errors.New("Couldn't inflate before: " + err.Error())
	}

	inflatedAfter, err := jd.ReadJsonString(string(after))
	if err != nil {
		return errors.New("Couldn't inflate after: " + err.Error())
	}

	diff, err := jsonBlobToDiff(formattedPatch)
	if err != nil {
		return errors.New("Couldn't inflate patch: " + err.Error())
	}

	patchedAfter, err := inflatedBefore.Patch(diff)
	if err != nil {
		return errors.New("Couldn't apply patch: " + err.Error())
	}

	if diff := deep.Equal(inflatedAfter, patchedAfter); len(diff) > 0 {
		return errors.New("The derived after didn't match the provided after: " + strings.Join(diff, "\n"))
	}

	return nil
}

func (j *josephBurnettEncoder) ApplyPatch(lastStateBlob, patchBlob []byte) (boardgame.StateStorageRecord, error) {
	lastStateNode, err := jd.ReadJsonString(string(lastStateBlob))

	if err != nil {
		return nil, errors.New("Couldn't parse lastStateBlob: " + err.Error())
	}

	diff, err := jsonBlobToDiff(patchBlob)

	if err != nil {
		return nil, errors.New("couldn unpack patch: " + err.Error())
	}

	result, err := lastStateNode.Patch(diff)

	if err != nil {
		return nil, errors.New("Couldn't patch diff: " + err.Error())
	}

	return []byte(result.Json()), nil
}

func (j *josephBurnettEncoder) Matches(examplePatch []byte) error {

	//WE match if they're strings, with the first line of the first file being "@"

	var probeStruct []string

	if err := json.Unmarshal(examplePatch, &probeStruct); err != nil {
		return errors.New("Unmarshal probe for unpacking patch failed " + err.Error())
	}

	if len(probeStruct) == 0 {
		return errors.New("No records")
	}

	if !strings.HasPrefix(probeStruct[0], "@") {
		return errors.New("The first record did not have the @ symbol as expected")
	}

	return nil

}
