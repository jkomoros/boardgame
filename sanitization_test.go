package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"
)

type testSanitizationDelegate struct {
	testGameDelegate
	policy *StatePolicy
}

func (t *testSanitizationDelegate) StateSanitizationPolicy() *StatePolicy {
	return t.policy
}

func TestSanitization(t *testing.T) {
	manager := NewGameManager(&testSanitizationDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.SetUp()

	tests := []struct {
		policy           *StatePolicy
		playerIndex      int
		inputFileName    string
		expectedFileName string
	}{
		{
			nil,
			-1,
			"sanitization_basic_in.json",
			"sanitization_basic_in.json",
		},
	}

	for i, test := range tests {

		inputBlob, err := ioutil.ReadFile("test/" + test.inputFileName)

		if err != nil {
			t.Fatal("couldn't load input file", i, test.inputFileName, err)
		}

		state, err := manager.Delegate().StateFromBlob(inputBlob)

		if err != nil {
			t.Fatal(i, "Failed to deserialize", err)
		}

		manager.delegate.(*testSanitizationDelegate).policy = test.policy

		sanitizedState := manager.SanitizedStateForPlayer(state, test.playerIndex)

		if sanitizedState == nil {
			t.Fatal(i, "state sanitization came back nil")
		}

		sanitizedBlob, err := json.Marshal(sanitizedState)

		if err != nil {
			t.Fatal(i, "Sanitized serialize failed", err)
		}

		goldenBlob, err := ioutil.ReadFile("test/" + test.expectedFileName)

		if err != nil {
			t.Fatal("Couldn't load file", i, test.expectedFileName, err)
		}

		compareJSONObjects(sanitizedBlob, goldenBlob, "Test Sanitization "+strconv.Itoa(i), t)

	}

}
