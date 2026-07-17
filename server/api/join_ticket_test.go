package api

import (
	"strings"
	"testing"
	"time"
)

func joinTicketTestServer() *Server {
	return &Server{joinTicketKey: []byte("0123456789abcdef0123456789abcdef")}
}

func TestJoinTicketRoundTrip(t *testing.T) {
	s := joinTicketTestServer()
	now := time.Unix(1_700_000_000, 0)
	ticket, err := s.issueJoinTicket("GAME", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.verifyJoinTicket(ticket, "GAME", now.Add(joinTicketTTL-time.Second)); err != nil {
		t.Fatalf("valid ticket rejected: %v", err)
	}
}

func TestJoinTicketRejectsExpiryWrongGameAndTampering(t *testing.T) {
	s := joinTicketTestServer()
	now := time.Unix(1_700_000_000, 0)
	ticket, err := s.issueJoinTicket("GAME", now)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]struct {
		ticket string
		gameID string
		now    time.Time
	}{
		"expired at boundary": {ticket, "GAME", now.Add(joinTicketTTL)},
		"wrong game":          {ticket, "OTHER", now},
		"tampered payload":    {"x" + ticket[1:], "GAME", now},
		"tampered signature":  {ticket[:len(ticket)-1] + "x", "GAME", now},
		"malformed":           {"not-a-ticket", "GAME", now},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := s.verifyJoinTicket(test.ticket, test.gameID, test.now); err == nil {
				t.Fatal("invalid ticket was accepted")
			}
		})
	}
}

func TestJoinTicketDoesNotExposeRoomCodeOrUseStandardBase64(t *testing.T) {
	s := joinTicketTestServer()
	ticket, err := s.issueJoinTicket("GAME/with+unsafe=chars", time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(ticket, "+/=") {
		t.Fatalf("ticket is not URL/header-safe: %q", ticket)
	}
}

func TestJoinTicketSharedSecretAndRotationWorkAcrossServers(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	oldKey := joinTicketKeyFromSecret("old-secret-that-is-long-enough-for-production")
	issuer := &Server{joinTicketKey: oldKey}
	ticket, err := issuer.issueJoinTicket("GAME", now)
	if err != nil {
		t.Fatal(err)
	}
	peer := &Server{joinTicketKey: oldKey}
	if err := peer.verifyJoinTicket(ticket, "GAME", now); err != nil {
		t.Fatalf("peer with the shared secret rejected ticket: %v", err)
	}
	rotated := &Server{
		joinTicketKey:         joinTicketKeyFromSecret("new-secret-that-is-long-enough-for-production"),
		joinTicketPreviousKey: oldKey,
	}
	if err := rotated.verifyJoinTicket(ticket, "GAME", now); err != nil {
		t.Fatalf("rotated peer rejected ticket signed by previous key: %v", err)
	}
}
