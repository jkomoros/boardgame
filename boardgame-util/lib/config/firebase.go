package config

type FirebaseConfig struct {
	ApiKey            string `json:"apiKey"`
	AuthDomain        string `json:"autoDomain"`
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
