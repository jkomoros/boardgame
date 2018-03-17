# go-firebase-verify
Verifies Backend Tokens from Firebase.

## Usage

`VerifyIDToken(idToken string, googleProjectID string) (string, error)`

Takes the JWT Token and googleProjectID as a parameter, returns the User ID (or an error).
