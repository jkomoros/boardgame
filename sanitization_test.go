package boardgame

import (
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

	game := NewGame(manager)

	game.SetUp(0)

	tests := []struct {
		policy           *StatePolicy
		expectedFileName string
	}{}

	for _, _ = range tests {
		//TODO: actually do the tests

	}

}
