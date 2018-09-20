package config

import (
	"encoding/json"
	"net/url"
	"strings"
)

//ConfigMode is the final, derived struct holding all of the leaf values in
//config.
type ConfigMode struct {
	//ConfigMode is primarily just the common config mode values
	ConfigModeCommon
	//Games is not intended to be inflated from JSON, but rather is derived
	//based on the contents of Games. It is OK to use literally as Games in
	//RawConfig, though, because its serialization is a legal GamesNode.
	Games []string `json:"games"`

	parentConfig *Config
}

func (c *ConfigMode) String() string {
	blob, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR, couldn't unmarshal: " + err.Error()
	}
	return string(blob)
}

//ParentConfig returns the Config that this ConfigMode is part of.
//Specifically, returns the config that was passed as ParentConfig to
//RawConfigMode.Derive().
func (c *ConfigMode) ParentConfig() *Config {
	return c.parentConfig
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
