package api

import (
	"encoding/json"
	"testing"
)

func TestNextAnimationPlayAtStartsWithLead(t *testing.T) {
	now := int64(10_000)
	got := nextAnimationPlayAt(now, 0)
	want := now + animationLeadMS
	if got != want {
		t.Fatalf("nextAnimationPlayAt(%d, 0) = %d, want %d", now, got, want)
	}
}

func TestNextAnimationPlayAtPacesBurst(t *testing.T) {
	first := nextAnimationPlayAt(10_000, 0)
	second := nextAnimationPlayAt(10_005, first)
	third := nextAnimationPlayAt(10_010, second)

	if second-first != companionAnimationSlotMS {
		t.Fatalf("second slot spacing = %dms, want %dms", second-first, companionAnimationSlotMS)
	}
	if third-second != companionAnimationSlotMS {
		t.Fatalf("third slot spacing = %dms, want %dms", third-second, companionAnimationSlotMS)
	}
}

func TestNextAnimationPlayAtRestartsAfterIdle(t *testing.T) {
	previous := int64(10_250)
	now := previous + companionAnimationSlotMS + 1_000
	got := nextAnimationPlayAt(now, previous)
	want := now + animationLeadMS
	if got != want {
		t.Fatalf("slot after idle = %d, want fresh lead %d", got, want)
	}
}

func TestMarshalVersionFramesUsesReservedTiming(t *testing.T) {
	versionData, timingData := marshalVersionFrames(42, 1_000, 2_345)

	var version socketMessage
	if err := json.Unmarshal(versionData, &version); err != nil {
		t.Fatal(err)
	}
	if version.Type != "version" || version.Data.(float64) != 42 {
		t.Fatalf("unexpected version frame: %#v", version)
	}

	var timing socketMessage
	if err := json.Unmarshal(timingData, &timing); err != nil {
		t.Fatal(err)
	}
	data := timing.Data.(map[string]interface{})
	if timing.Type != "version-timing" || data["version"].(float64) != 42 ||
		data["serverSentAt"].(float64) != 1_000 || data["serverPlayAt"].(float64) != 2_345 ||
		data["slotDurationMs"].(float64) != companionAnimationSlotMS ||
		data["maxAnimationDurationMs"].(float64) != companionAnimationDurationMS {
		t.Fatalf("unexpected timing frame: %#v", timing)
	}
}

func TestRegisterSocketReusesCurrentVersionLaneTiming(t *testing.T) {
	v := newTestNotifier(t)
	v.animationLaneTail = map[string]animationLaneEntry{
		"game": {version: 42, serverPlayAt: 4_500},
	}
	s := &socket{gameID: "game", initialVersion: 42, send: make(chan socketFrameBatch, 1)}
	v.registerSocket(s)

	batch := <-s.send
	if len(batch) != 2 {
		t.Fatalf("register batch has %d frames, want version + timing", len(batch))
	}
	var timing socketMessage
	if err := json.Unmarshal(batch[1], &timing); err != nil {
		t.Fatal(err)
	}
	data := timing.Data.(map[string]interface{})
	if data["serverPlayAt"].(float64) != 4_500 {
		t.Fatalf("register timing did not reuse lane tail: %#v", timing)
	}
}

func TestRegisterSocketCatchesUpPastHandshakeSnapshot(t *testing.T) {
	v := newTestNotifier(t)
	v.animationLaneTail["game"] = animationLaneEntry{version: 43, serverPlayAt: 4_500}
	s := &socket{gameID: "game", initialVersion: 42, send: make(chan socketFrameBatch, 1)}
	v.registerSocket(s)

	batch := <-s.send
	var version socketMessage
	if err := json.Unmarshal(batch[0], &version); err != nil {
		t.Fatal(err)
	}
	var timing socketMessage
	if err := json.Unmarshal(batch[1], &timing); err != nil {
		t.Fatal(err)
	}
	data := timing.Data.(map[string]interface{})
	if version.Data.(float64) != 43 || data["version"].(float64) != 43 ||
		data["serverPlayAt"].(float64) != 4_500 {
		t.Fatalf("register sent stale handshake version: version=%#v timing=%#v", version, timing)
	}
}

func TestNotificationWithoutListenersStillCatchesUpRegistration(t *testing.T) {
	v := newTestNotifier(t)
	go v.workLoop()
	t.Cleanup(func() { v.done() })

	// Both channels are unbuffered. Once the registration send is accepted,
	// workLoop has necessarily finished reserving the preceding notification.
	for version := 43; version <= 142; version++ {
		v.notifyVersion <- gameVersionChanged{ID: "game", Version: version}
	}
	s := &socket{gameID: "game", initialVersion: 42, send: make(chan socketFrameBatch, 1)}
	v.register <- s
	batch := <-s.send

	var version socketMessage
	if err := json.Unmarshal(batch[0], &version); err != nil {
		t.Fatal(err)
	}
	if version.Data.(float64) != 142 {
		t.Fatalf("register sent version %v, want latest notification's 142", version.Data)
	}
	var timing socketMessage
	if err := json.Unmarshal(batch[1], &timing); err != nil {
		t.Fatal(err)
	}
	data := timing.Data.(map[string]interface{})
	lead := data["serverPlayAt"].(float64) - data["serverSentAt"].(float64)
	if lead > animationLeadMS+100 {
		t.Fatalf("listenerless burst built a future backlog: lead=%vms", lead)
	}
}

func TestSendMessageEnqueuesVersionAndTimingAtomically(t *testing.T) {
	s := &socket{send: make(chan socketFrameBatch, 1)}
	s.SendMessage([]byte("version"), []byte("timing"))
	batch := <-s.send
	if len(batch) != 2 || string(batch[0]) != "version" || string(batch[1]) != "timing" {
		t.Fatalf("unexpected frame batch: %#v", batch)
	}
}

func TestReserveAnimationLaneIsIdempotentAndRejectsRegression(t *testing.T) {
	first, ok := reserveAnimationLane(1_000, 10, animationLaneEntry{}, false, true)
	if !ok {
		t.Fatal("first reservation unexpectedly rejected")
	}
	duplicate, ok := reserveAnimationLane(1_100, 10, first, true, true)
	if !ok || duplicate != first {
		t.Fatalf("duplicate changed lane: first=%#v duplicate=%#v ok=%v", first, duplicate, ok)
	}
	regressed, ok := reserveAnimationLane(1_200, 9, first, true, true)
	if ok || regressed != first {
		t.Fatalf("regression was not rejected: first=%#v got=%#v ok=%v", first, regressed, ok)
	}
}

func TestUnobservedVersionsCoalesceInsteadOfBuildingBacklog(t *testing.T) {
	var tail animationLaneEntry
	hasTail := false
	for version := 1; version <= 100; version++ {
		var ok bool
		tail, ok = reserveAnimationLane(1_000+int64(version), version, tail, hasTail, false)
		if !ok {
			t.Fatalf("version %d unexpectedly rejected", version)
		}
		hasTail = true
	}
	want := int64(1_100 + animationLeadMS)
	if tail.serverPlayAt != want {
		t.Fatalf("unobserved burst target = %d, want coalesced %d", tail.serverPlayAt, want)
	}
}

func TestRegistrationBaselineDoesNotConsumeNextMoveSlot(t *testing.T) {
	baseline, ok := reserveAnimationLane(1_000, 10, animationLaneEntry{}, false, false)
	if !ok {
		t.Fatal("baseline reservation unexpectedly rejected")
	}
	baseline.pacesNext = false
	next, ok := reserveAnimationLane(1_050, 11, baseline, true, true)
	if !ok {
		t.Fatal("next move unexpectedly rejected")
	}
	if next.serverPlayAt != 1_050+animationLeadMS {
		t.Fatalf("first move target = %d, want fresh lead", next.serverPlayAt)
	}
}

func TestSendSocketMessageCreatesOneFrameBatch(t *testing.T) {
	s := &socket{send: make(chan socketFrameBatch, 1)}
	s.SendSocketMessage("clock-sync", map[string]interface{}{"clientSentAt": 123})
	batch := <-s.send
	if len(batch) != 1 {
		t.Fatalf("clock sync batch has %d frames, want 1", len(batch))
	}
	var msg socketMessage
	if err := json.Unmarshal(batch[0], &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "clock-sync" {
		t.Fatalf("unexpected message: %#v", msg)
	}
}

func TestModeChangedFrameCarriesGameIdentity(t *testing.T) {
	v := newTestNotifier(t)
	s := &socket{gameID: "game", send: make(chan socketFrameBatch, 1)}
	v.sockets["game"] = map[*socket]bool{s: true}

	v.doBroadcastModeChanged("game", "solo")

	batch := <-s.send
	if len(batch) != 1 {
		t.Fatalf("mode change batch has %d frames, want 1", len(batch))
	}
	var msg socketMessage
	if err := json.Unmarshal(batch[0], &msg); err != nil {
		t.Fatal(err)
	}
	data := msg.Data.(map[string]interface{})
	if msg.Type != "mode-changed" || data["gameID"] != "game" || data["newMode"] != "solo" {
		t.Fatalf("unexpected mode change frame: %#v", msg)
	}
}
