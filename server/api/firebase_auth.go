package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
)

const (
	firebaseVerifyTimeout         = 2 * time.Second
	firebaseInitializationTimeout = 5 * time.Second
)

var errFirebaseVerifyTimeout = errors.New("Firebase token verification timed out")

type firebaseTokenVerifier interface {
	VerifyIDToken(context.Context, string) (*firebaseauth.Token, error)
}

func newFirebaseTokenVerifier(ctx context.Context, projectID string) (firebaseTokenVerifier, error) {
	if projectID == "" {
		return nil, errors.New("Firebase project ID is required")
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Auth client: %w", err)
	}
	return client, nil
}

func verifyFirebaseTokenWithTimeout(
	ctx context.Context,
	verifier firebaseTokenVerifier,
	token string,
	timeout time.Duration,
) (string, error) {
	if verifier == nil {
		return "", errors.New("Firebase token verifier is not initialized")
	}
	verifyCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	decoded, err := verifier.VerifyIDToken(verifyCtx, token)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(verifyCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: %v", errFirebaseVerifyTimeout, err)
		}
		return "", err
	}
	if decoded == nil || decoded.UID == "" {
		return "", errors.New("Firebase token did not contain a UID")
	}
	return decoded.UID, nil
}
