package boardgame

import (
	"testing"
)

func TestChatConfigBuilder(t *testing.T) {
	// Default: everything enabled
	c := DefaultChatConfig()
	if !c.IsEnabled() {
		t.Error("Default should be enabled")
	}
	if !c.AllChatEnabled() {
		t.Error("Default should have all-chat")
	}
	if !c.TeamChatEnabled() {
		t.Error("Default should have team chat")
	}
	if !c.DMChatEnabled() {
		t.Error("Default should have DM chat")
	}
	if c.IsPrebakedOnly() {
		t.Error("Default should not be prebaked-only")
	}

	// Disabled
	c2 := DefaultChatConfig().Disabled()
	if c2.IsEnabled() {
		t.Error("Disabled should not be enabled")
	}
	if c2.AllChatEnabled() {
		t.Error("Disabled should not have all-chat")
	}

	// WithoutDMs
	c3 := DefaultChatConfig().WithoutDMs()
	if !c3.IsEnabled() {
		t.Error("WithoutDMs should still be enabled")
	}
	if !c3.AllChatEnabled() {
		t.Error("WithoutDMs should still have all-chat")
	}
	if c3.DMChatEnabled() {
		t.Error("WithoutDMs should not have DM chat")
	}

	// WithoutTeamChat
	c4 := DefaultChatConfig().WithoutTeamChat()
	if c4.TeamChatEnabled() {
		t.Error("WithoutTeamChat should not have team chat")
	}
	if !c4.DMChatEnabled() {
		t.Error("WithoutTeamChat should still have DM chat")
	}

	// PrebakedOnly
	c5 := DefaultChatConfig().PrebakedOnly("👍", "😂", "gg")
	if !c5.IsPrebakedOnly() {
		t.Error("PrebakedOnly should be prebaked")
	}
	if len(c5.AllowedMessages()) != 3 {
		t.Errorf("PrebakedOnly should have 3 messages, got %d", len(c5.AllowedMessages()))
	}

	// Chaining
	c6 := DefaultChatConfig().WithoutDMs().WithoutTeamChat().PrebakedOnly("yes", "no")
	if c6.DMChatEnabled() {
		t.Error("Chained should not have DMs")
	}
	if c6.TeamChatEnabled() {
		t.Error("Chained should not have team chat")
	}
	if !c6.IsPrebakedOnly() {
		t.Error("Chained should be prebaked")
	}
	if !c6.AllChatEnabled() {
		t.Error("Chained should still have all-chat")
	}
}

func TestChatPolicyStruct(t *testing.T) {
	p := ChatPolicy{
		Enabled:      true,
		SendChannels: []string{"all", "team/Red"},
		ViewChannels: []string{"all", "team/Red", "dm/a/b"},
	}
	if !p.Enabled {
		t.Error("Policy should be enabled")
	}
	if len(p.SendChannels) != 2 {
		t.Errorf("Expected 2 send channels, got %d", len(p.SendChannels))
	}
	if len(p.ViewChannels) != 3 {
		t.Errorf("Expected 3 view channels, got %d", len(p.ViewChannels))
	}
}
