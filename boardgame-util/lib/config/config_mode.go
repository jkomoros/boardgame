package config

import (
	"errors"
	"log"
	"net/url"
	"strings"
)

//ConfigMode is the final, derived struct holding all of the leaf values in
//config.
type ConfigMode struct {
	//ConfigMode is primarily just the common config mode values
	ConfigModeCommon
	//GamesList is not intended to be inflated from JSON, but rather is
	//derived based on the contents of Games.
	GamesList []string
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
