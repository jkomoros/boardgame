package api

import (
	"strings"
	"testing"
	"time"

	"github.com/jkomoros/boardgame/server/api/tablelease"
)

func TestTableLeaseCredentialRoundTripAndTampering(t *testing.T) {
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	deviceID, secret, digest, err := s.newTableLeaseCredential("GAME")
	if err != nil {
		t.Fatal(err)
	}
	credential := deviceID + "." + secret
	record := &tablelease.StorageRecord{DeviceID: deviceID, SecretDigest: digest}
	if !tableLeaseCredentialMatches(record, credential) {
		t.Fatal("fresh credential did not match its digest")
	}
	tests := []string{
		"", credential + "x", "not.hex", strings.Repeat("x", 257),
		"0" + credential[1:], credential[:len(credential)-1] + "0",
	}
	for _, invalid := range tests {
		if tableLeaseCredentialMatches(record, invalid) {
			t.Fatalf("tampered credential unexpectedly matched: %q", invalid)
		}
	}
}

func TestTableLeaseCredentialIsStableForIdempotentDeviceRetries(t *testing.T) {
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	const deviceID = "0123456789abcdef0123456789abcdef"
	first, firstDigest, err := s.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	second, secondDigest, err := s.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || firstDigest != secondDigest {
		t.Fatal("same game/device retry rotated its credential")
	}
	other, _, err := s.tableLeaseCredentialForDevice("OTHER", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if first == other {
		t.Fatal("credential was not bound to its game")
	}
}

func TestTableLeaseCredentialIgnoresJoinTicketRotationAcrossServers(t *testing.T) {
	sharedTableKey := []byte("table-key-shared-by-every-api-instance")
	oldServer := &Server{tableLeaseKey: sharedTableKey, joinTicketKey: []byte("old-join-key-old-join-key-old-key")}
	newServer := &Server{tableLeaseKey: sharedTableKey, joinTicketKey: []byte("new-join-key-new-join-key-new-key")}
	const deviceID = "fedcba9876543210fedcba9876543210"
	oldCredential, _, err := oldServer.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	newCredential, _, err := newServer.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if oldCredential != newCredential {
		t.Fatal("join-ticket rotation changed a Table recovery credential")
	}
}

func TestTableLeaseActiveRequiresCredentialAndFutureExpiry(t *testing.T) {
	now := time.UnixMilli(10_000)
	record := &tablelease.StorageRecord{DeviceID: "device", SecretDigest: "digest", Expires: 10_001}
	if !tableLeaseActive(record, now) {
		t.Fatal("future complete lease should be active")
	}
	record.Expires = 10_000
	if tableLeaseActive(record, now) {
		t.Fatal("lease must expire at the exact boundary")
	}
	record.Expires = 10_001
	record.SecretDigest = ""
	if tableLeaseActive(record, now) {
		t.Fatal("lease without a secret digest must be inactive")
	}
}
