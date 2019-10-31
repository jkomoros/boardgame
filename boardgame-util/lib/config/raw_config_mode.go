package config

import (
	"encoding/json"
)

//RawConfigMode is the leaf of RawConfig, where all of the actual values are
//stored.
type RawConfigMode struct {
	//ConfigMode is primarily just the common config mode values
	ModeCommon
	Games *GameNode `json:"games,omitempty"`
}

//Derive tells the RawConfigMode to create a new, fully derived ConfigMode
//based on the current properties of this RawConfigMode, setting defaults as
//necessary. parentConfig is the Config that this configMode will be part of.
//prodMode is whether the ConfigMode being derived is for a Prod or Dev slot
//in Config. Will always return a reasonably defaulted ConfigMode even if the
//RawcConfigMode itself is nil. Generally you don't call this, but use
//NewConfig() instead.
func (c *RawConfigMode) Derive(parentConfig *Config, prodMode bool) *Mode {

	var result *Mode

	if c == nil {
		result = &Mode{}
	} else {
		result = &Mode{
			ModeCommon: c.ModeCommon,
			Games:      c.Games.List(),
		}
	}

	result.parentConfig = parentConfig

	if result.DefaultPort == "" {
		if prodMode {
			result.DefaultPort = "8080"
		} else {
			result.DefaultPort = "8888"
		}
	}

	if result.DefaultStaticPort == "" {
		if prodMode {
			result.DefaultStaticPort = "80"
		} else {
			result.DefaultStaticPort = "8080"
		}
	}

	//AllowedOrigins will just be default allow
	if result.AllowedOrigins == "" {
		result.AllowedOrigins = "*"
	}

	if result.APIHost == "" {
		if prodMode {
			if result.Firebase != nil {
				result.APIHost = "https://" + result.Firebase.StorageBucket
			}
		} else {
			result.APIHost = "http://localhost"
		}
		if result.APIHost != "" {
			if result.DefaultPort != "80" && result.DefaultPort != "" {
				result.APIHost += ":" + result.DefaultPort
			}
		}
	}

	if result.Storage == nil {
		result.Storage = make(map[string]string)
	}

	if result.DisableAdminChecking && prodMode {
		//Not legal, turn off!

		//TODO: ideally we'd communicate that we had unset this...
		result.DisableAdminChecking = false
	}

	if result.OfflineDevMode && prodMode {
		//Not legal, turn off!

		//TODO: ideally we'd communicate that we had unset this...
		result.OfflineDevMode = false
	}

	return result

}

func (c *RawConfigMode) String() string {
	blob, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR, couldn't unmarshal: " + err.Error()
	}
	return string(blob)
}

//Copy returns a deep copy of the RawConfigMode.
func (c *RawConfigMode) Copy() *RawConfigMode {

	if c == nil {
		return nil
	}

	result := &RawConfigMode{}

	(*result) = *c
	result.AdminUserIds = make([]string, len(c.AdminUserIds))
	copy(result.AdminUserIds, c.AdminUserIds)
	result.Storage = make(map[string]string, len(c.Storage))
	for key, val := range c.Storage {
		result.Storage[key] = val
	}

	result.Games = result.Games.copy()
	result.Firebase = result.Firebase.copy()

	return result

}

//mergedStrList returns a list where base is concatenated with the non-
//duplicates in other.
func mergedStrList(base, other []string) []string {

	if base == nil {
		return other
	}

	if other == nil {
		return base
	}

	result := make([]string, len(base))

	for i, str := range base {
		result[i] = str
	}

	strSet := make(map[string]bool, len(base))
	for _, key := range base {
		strSet[key] = true
	}

	for _, key := range other {
		if strSet[key] {
			//Already in the set, don't add a duplicate
			continue
		}
		result = append(result, key)
	}

	return result
}

//Extend takes a given base config mode, extends it with properties set in
//other (with any non-zero value overwriting the base values, and with Games
//and string lists being merged and de-duped) and returns a *new* config
//representing the merged one. Normally you don't call this directly but use
//NewConfig instead.
func (c *RawConfigMode) Extend(other *RawConfigMode) *RawConfigMode {

	if c == nil && other != nil {
		return other.Copy()
	}

	result := c.Copy()

	if other == nil {
		return result
	}

	if other.AllowedOrigins != "" {
		result.AllowedOrigins = other.AllowedOrigins
	}

	if other.DefaultPort != "" {
		result.DefaultPort = other.DefaultPort
	}

	if other.DefaultStaticPort != "" {
		result.DefaultStaticPort = other.DefaultStaticPort
	}

	if other.DisableAdminChecking {
		result.DisableAdminChecking = true
	}

	if other.OfflineDevMode {
		result.OfflineDevMode = true
	}

	if other.GoogleAnalytics != "" {
		result.GoogleAnalytics = other.GoogleAnalytics
	}

	if other.APIHost != "" {
		result.APIHost = other.APIHost
	}

	result.AdminUserIds = mergedStrList(c.AdminUserIds, other.AdminUserIds)

	for key, val := range other.Storage {
		result.Storage[key] = val
	}

	result.Games = result.Games.extend(other.Games)

	result.Firebase = result.Firebase.extend(other.Firebase)

	return result

}
