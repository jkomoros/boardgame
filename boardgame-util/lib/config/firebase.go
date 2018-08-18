package config

import (
	"strings"
)

//FirebaseKey denotes a specific key in a firebase object
type FirebaseKey string

const (
	FirebaseInvalid           FirebaseKey = "<INVALID>"
	FirebaseApiKey                        = "apiKey"
	FirebaseAuthDomain                    = "authDomain"
	FirebaseDatabaseURL                   = "databaseURL"
	FirebaseProjectID                     = "projectId"
	FirebaseStorageBucket                 = "storageBucket"
	FirebaseMessagingSenderID             = "messagingSenderId"
)

//FirebaseKeys enumerates all firebase keys
var FirebaseKeys = map[FirebaseKey]bool{
	FirebaseApiKey:            true,
	FirebaseAuthDomain:        true,
	FirebaseDatabaseURL:       true,
	FirebaseProjectID:         true,
	FirebaseStorageBucket:     true,
	FirebaseMessagingSenderID: true,
}

//FirebaseConfig is a sub-struct within ConfigMode that holds values specific
//to firebase.
type FirebaseConfig struct {
	ApiKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	DatabaseURL       string `json:"databaseURL"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
}

func (f *FirebaseConfig) copy() *FirebaseConfig {

	if f == nil {
		return nil
	}

	result := &FirebaseConfig{}
	(*result) = *f
	return result
}

func (f *FirebaseConfig) extend(other *FirebaseConfig) *FirebaseConfig {
	if f == nil {
		return nil
	}
	result := f.copy()

	if other == nil {
		return result
	}

	if other.ApiKey != "" {
		result.ApiKey = other.ApiKey
	}

	if other.AuthDomain != "" {
		result.AuthDomain = other.AuthDomain
	}

	if other.DatabaseURL != "" {
		result.DatabaseURL = other.DatabaseURL
	}

	if other.ProjectID != "" {
		result.ProjectID = other.ProjectID
	}

	if other.StorageBucket != "" {
		result.StorageBucket = other.StorageBucket
	}

	if other.MessagingSenderID != "" {
		result.MessagingSenderID = other.MessagingSenderID
	}

	return result
}

//FirebaseKeyFromString returns the FirebseKey denoted by key (fuzzy
//matching).
func FirebaseKeyFromString(key string) FirebaseKey {

	key = strings.ToLower(key)
	key = strings.TrimSpace(key)

	for name, _ := range FirebaseKeys {
		normalizedName := strings.ToLower(string(name))

		if normalizedName == key {
			return name
		}
	}

	return FirebaseInvalid

}
