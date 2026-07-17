package tablelease

import "testing"

const (
	testTransferID     = "0123456789abcdef0123456789abcdef"
	testTransferDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	testDeviceID       = "abcdef0123456789abcdef0123456789"
)

func TestValidateTransferAcceptsEmptyPendingAndRedeemedStates(t *testing.T) {
	valid := []*StorageRecord{
		{},
		{
			DeviceID:            testDeviceID,
			TransferID:          testTransferID,
			TransferTokenDigest: testTransferDigest,
			TransferCodeDigest:  testTransferDigest,
			TransferExpires:     1,
		},
		{DeviceID: testDeviceID, PreviousDeviceID: testDeviceID, TransitionKind: TransitionSolo},
		{
			DeviceID: testDeviceID, PreviousDeviceID: testDeviceID, TransitionKind: TransitionHostAction,
			TransferID: testTransferID, TransferTokenDigest: testTransferDigest,
			TransferCodeDigest: testTransferDigest, TransferExpires: 1,
		},
		{
			DeviceID:               testDeviceID,
			TransferID:             testTransferID,
			TransferTokenDigest:    testTransferDigest,
			TransferCodeDigest:     testTransferDigest,
			TransferExpires:        1,
			TransferTargetDeviceID: testDeviceID,
			PreviousDeviceID:       testTransferID,
			TransitionKind:         TransitionTransfer,
		},
		{
			DeviceID: testDeviceID, TransferID: testTransferID,
			TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest, TransferExpires: 1,
			TransferTargetDeviceID: testDeviceID, PreviousDeviceID: testTransferID,
			TransitionKind: TransitionHostAction,
		},
	}
	for index, record := range valid {
		if err := record.ValidateTransfer(); err != nil {
			t.Fatalf("valid record %d was rejected: %v", index, err)
		}
	}
}

func TestValidateTransferRejectsPartialOrMalformedStates(t *testing.T) {
	invalid := []*StorageRecord{
		{TransferID: testTransferID},
		{TransferID: testTransferID, TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest},
		{TransferID: "ABCDEF0123456789ABCDEF0123456789", TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest, TransferExpires: 1},
		{TransferID: testTransferID, TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest, TransferExpires: 1, TransferTargetDeviceID: "short"},
		{PreviousDeviceID: testDeviceID, TransitionKind: "unknown"},
		{TransitionKind: TransitionRecovery},
		{DeviceID: testDeviceID, PreviousDeviceID: testTransferID, TransitionKind: TransitionHostAction},
		{DeviceID: testDeviceID, PreviousDeviceID: testDeviceID, TransitionKind: TransitionRecovery},
		{DeviceID: testDeviceID, TransferID: testTransferID, TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest, TransferExpires: 1, TransferTargetDeviceID: testTransferID, PreviousDeviceID: testTransferID, TransitionKind: TransitionTransfer},
		{DeviceID: testDeviceID, TransferID: testTransferID, TransferTokenDigest: testTransferDigest, TransferCodeDigest: testTransferDigest, TransferExpires: 1, TransferTargetDeviceID: testDeviceID},
	}
	for index, record := range invalid {
		if err := record.ValidateTransfer(); err == nil {
			t.Fatalf("invalid record %d was accepted: %+v", index, record)
		}
	}
}

func TestTransferStateAndClear(t *testing.T) {
	record := &StorageRecord{
		DeviceID:            testDeviceID,
		TransferID:          testTransferID,
		TransferTokenDigest: testTransferDigest,
		TransferCodeDigest:  testTransferDigest,
		TransferExpires:     100,
	}
	if !record.TransferPending(99) || record.TransferPending(100) || record.TransferRedeemed(99) {
		t.Fatal("pending transfer state did not respect target and expiry")
	}
	record.TransferTargetDeviceID = testDeviceID
	record.PreviousDeviceID = testTransferID
	record.TransitionKind = TransitionTransfer
	if record.TransferPending(99) || !record.TransferRedeemed(99) || record.TransferRedeemed(100) {
		t.Fatal("redeemed transfer state did not respect target and expiry")
	}
	record.ClearTransfer()
	if record.DeviceID != testDeviceID || record.TransferID != "" || record.TransferTokenDigest != "" ||
		record.TransferCodeDigest != "" || record.TransferExpires != 0 || record.TransferTargetDeviceID != "" {
		t.Fatalf("clear did not isolate transfer fields: %+v", record)
	}
}
