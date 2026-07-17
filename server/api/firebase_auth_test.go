package api

import (
	"context"
	"errors"
	"testing"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
)

type fakeFirebaseTokenVerifier func(context.Context, string) (*firebaseauth.Token, error)

func (f fakeFirebaseTokenVerifier) VerifyIDToken(ctx context.Context, token string) (*firebaseauth.Token, error) {
	return f(ctx, token)
}

func TestNewFirebaseTokenVerifier(t *testing.T) {
	t.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "127.0.0.1:9099")
	verifier, err := newFirebaseTokenVerifier(context.Background(), "test-project")
	if err != nil || verifier == nil {
		t.Fatalf("emulator verifier = (%v, %v), want non-nil verifier", verifier, err)
	}
	if _, err := newFirebaseTokenVerifier(context.Background(), ""); err == nil {
		t.Fatal("empty Firebase project ID was accepted")
	}
}

func TestVerifyFirebaseTokenWithTimeout(t *testing.T) {
	verifier := fakeFirebaseTokenVerifier(func(_ context.Context, token string) (*firebaseauth.Token, error) {
		if token != "valid" {
			return nil, errors.New("invalid token")
		}
		return &firebaseauth.Token{UID: "player"}, nil
	})
	uid, err := verifyFirebaseTokenWithTimeout(context.Background(), verifier, "valid", time.Second)
	if err != nil || uid != "player" {
		t.Fatalf("verification = (%q, %v), want (player, nil)", uid, err)
	}
	if _, err := verifyFirebaseTokenWithTimeout(context.Background(), verifier, "invalid", time.Second); err == nil {
		t.Fatal("invalid token was accepted")
	}
	if _, err := verifyFirebaseTokenWithTimeout(context.Background(), nil, "valid", time.Second); err == nil {
		t.Fatal("nil verifier was accepted")
	}
}

func TestVerifyFirebaseTokenTimeout(t *testing.T) {
	verifier := fakeFirebaseTokenVerifier(func(ctx context.Context, _ string) (*firebaseauth.Token, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	_, err := verifyFirebaseTokenWithTimeout(context.Background(), verifier, "token", time.Millisecond)
	if !errors.Is(err, errFirebaseVerifyTimeout) {
		t.Fatalf("timeout error = %v, want errFirebaseVerifyTimeout", err)
	}
}
