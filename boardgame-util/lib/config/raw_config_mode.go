package config

import (
	"encoding/json"
)

type ConfigModeField string

const (
	FieldAllowedOrigins       ConfigModeField = "AllowedOrigins"
	FieldDefaultPort                          = "DefaultPort"
	FieldDefaultStaticPort                    = "DefaultStaticPort"
	FieldAdminUserIds                         = "AdminUserIds"
	FieldDisableAdminChecking                 = "DisableAdminChecking"
	FieldStorageConfig                        = "StorageConfig"
	FieldDefaultStorageType                   = "DefaultStorageType"
	FieldGoogleAnalytics                      = "GoogleAnalytics"
	FieldFirebase                             = "Firebase"
	FieldApiHost                              = "ApiHost"
	FieldGames                                = "Games"
)

//ConfigModeCommon is the values that both ConfigMode and RawConfigMode share
//directly, factored out for convenience so they can be anonymously embedded
//in ConfigMdoe and RawConfigMode.
type ConfigModeCommon struct {
	AllowedOrigins    string   `json:"allowedOrigins,omitempty"`
	DefaultPort       string   `json:"defaultPort,omitempty"`
	DefaultStaticPort string   `json:"defaultStaticPort,omitempty"`
	AdminUserIds      []string `json:"adminUserIds,omitempty"`
	//This is a dangerous config. Only enable in Dev!
	DisableAdminChecking bool              `json:"disableAdminChecking,omitempty"`
	StorageConfig        map[string]string `json:"storageConfig,omitempty"`
	//The storage type that should be used if no storage type is provided via
	//command line options.
	DefaultStorageType string `json:"defaultStorageType,omitempty"`

	//The GA config string. Will be used to generate the client_config json
	//blob. Generally has a structure like "UA-321655-11"
	GoogleAnalytics string          `json:"googleAnalytics,omitempty"`
	Firebase        *FirebaseConfig `json:"firebase,omitempty"`
	//The host name the client should connect to in that mode. Something like
	//"http://localhost:8888"
	ApiHost string `json:"apiHost,omitempty"`
}

//RawConfigMode is the leaf of RawConfig, where all of the actual values are
//stored.
type RawConfigMode struct {
	//ConfigMode is primarily just the common config mode values
	ConfigModeCommon
	Games *GameNode `json:"games,omitempty"`
}

//Derive tells the RawConfigMode to create a new, fully derived ConfigMode
//based on the current properties of this RawConfigMode, setting defaults as
//necessary. prodMode is whether the ConfigMode being derived is for a Prod or
//Dev slot in Config. Will always return a reasonably defaulted ConfigMode
//even if the RawcConfigMode itself is nil. Generally you don't call this, but
//use NewConfig() instead.
func (c *RawConfigMode) Derive(prodMode bool) *ConfigMode {

	var result *ConfigMode

	if c == nil {
		result = &ConfigMode{}
	} else {
		result = &ConfigMode{
			c.ConfigModeCommon,
			c.Games.List(),
		}
	}

	if result.DefaultPort == "" {
		if prodMode {
			result.DefaultPort = "8080"
		} else {
			result.DefaultPort = "8888"
		}
	}
	//AllowedOrigins will just be default allow
	if result.AllowedOrigins == "" {
		result.AllowedOrigins = "*"
	}

	if result.ApiHost == "" {
		if prodMode {
			if result.Firebase != nil {
				result.ApiHost = "https://" + result.Firebase.StorageBucket
			}
		} else {
			result.ApiHost = "http://localhost"
		}
		if result.ApiHost != "" {
			if result.DefaultPort != "80" && result.DefaultPort != "" {
				result.ApiHost += ":" + result.DefaultPort
			}
		}
	}

	if result.StorageConfig == nil {
		result.StorageConfig = make(map[string]string)
	}

	if result.DisableAdminChecking && prodMode {
		//Not legal, turn off!

		//TODO: ideally we'd communicate that we had unset this...
		result.DisableAdminChecking = false
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
	result.StorageConfig = make(map[string]string, len(c.StorageConfig))
	for key, val := range c.StorageConfig {
		result.StorageConfig[key] = val
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

	if other.GoogleAnalytics != "" {
		result.GoogleAnalytics = other.GoogleAnalytics
	}

	if other.ApiHost != "" {
		result.ApiHost = other.ApiHost
	}

	result.AdminUserIds = mergedStrList(c.AdminUserIds, other.AdminUserIds)

	for key, val := range other.StorageConfig {
		result.StorageConfig[key] = val
	}

	result.Games = result.Games.extend(other.Games)
	result.Firebase = result.Firebase.extend(other.Firebase)

	return result

}
