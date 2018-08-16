package config

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strings"
)

type ConfigMode struct {
	AllowedOrigins    string
	DefaultPort       string
	DefaultStaticPort string
	AdminUserIds      []string
	//This is a dangerous config. Only enable in Dev!
	DisableAdminChecking bool
	StorageConfig        map[string]string
	//The storage type that should be used if no storage type is provided via
	//command line options.
	DefaultStorageType string

	//The GA config string. Will be used to generate the client_config json
	//blob. Generally has a structure like "UA-321655-11"
	GoogleAnalytics string
	Firebase        *FirebaseConfig
	//The host name the client should connect to in that mode. Something like
	//"http://localhost:8888"
	ApiHost string

	Games *GameNode
	//GamesList is not intended to be inflated from JSON, but rather is
	//derived based on the contents of Games.
	GamesList []string
}

//derive tells the ConfigMode to do its final processing to create any derived
//fields, like GamesList.
func (c *ConfigMode) derive(prodMode bool) {

	if c == nil {
		return
	}

	c.GamesList = c.Games.List()

	if c.ApiHost == "" {
		if prodMode {
			if c.Firebase == nil {
				return
			}
			c.ApiHost = "https://" + c.Firebase.StorageBucket

		} else {
			c.ApiHost = "http://localhost"
		}

		if c.DefaultPort != "80" && c.DefaultPort != "" {
			c.ApiHost += ":" + c.DefaultPort
		}
	}

}

func (c *ConfigMode) String() string {
	blob, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR, couldn't unmarshal: " + err.Error()
	}
	return string(blob)
}

func (c *ConfigMode) validate(isDev bool) error {
	if c.DefaultPort == "" {
		return errors.New("No default port provided")
	}
	//AllowedOrigins will just be default allow
	if c.AllowedOrigins == "" {
		log.Println("No AllowedOrigins found. Defaulting to '*'")
		c.AllowedOrigins = "*"
	}
	if c.StorageConfig == nil {
		c.StorageConfig = make(map[string]string)
	}
	if c.DisableAdminChecking && !isDev {
		return errors.New("DisableAdminChecking enabled in prod, which is illegal")
	}
	return nil
}

//copy returns a deep copy of the config mode.
func (c *ConfigMode) copy() *ConfigMode {

	if c == nil {
		return nil
	}

	result := &ConfigMode{}

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

//extend takes a given base config mode, extends it with properties set in
//other (with any non-zero value overwriting the base values) and returns a
//*new* config representing the merged one.
func (c *ConfigMode) extend(other *ConfigMode) *ConfigMode {

	result := c.copy()

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

func (c *ConfigMode) OriginAllowed(origin string) bool {

	originUrl, err := url.Parse(origin)

	if err != nil {
		return false
	}

	if c.AllowedOrigins == "" {
		return false
	}
	if c.AllowedOrigins == "*" {
		return true
	}
	allowedOrigins := strings.Split(c.AllowedOrigins, ",")
	for _, allowedOrigin := range allowedOrigins {
		u, err := url.Parse(allowedOrigin)

		if err != nil {
			continue
		}

		if u.Scheme == originUrl.Scheme && u.Host == originUrl.Host {
			return true
		}
	}
	return false
}
