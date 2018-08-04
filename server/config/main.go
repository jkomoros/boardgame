/*

Config is a simple library that manages config set-up for servers based on a config file.

*/
package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Base *ConfigMode
	Dev  *ConfigMode
	Prod *ConfigMode
}

type ConfigMode struct {
	AllowedOrigins    string
	DefaultPort       string
	FirebaseProjectId string
	AdminUserIds      []string
	//This is a dangerous config. Only enable in Dev!
	DisableAdminChecking bool
	StorageConfig        map[string]string
}

//derive takes a raw input and creates a struct with fully derived values in
//Dev/Prod ready for use.
func (c *Config) derive() {
	if c.Base == nil {
		return
	}

	c.Prod = c.Base.extend(c.Prod)
	c.Dev = c.Base.extend(c.Dev)

}

func (c *Config) copy() *Config {
	return &Config{
		c.Base.copy(),
		c.Dev.copy(),
		c.Prod.copy(),
	}
}

//extend takes an other config and returns a *new* config where any non-zero
//value for other extends base.
func (c *Config) extend(other *Config) *Config {

	result := c.copy()

	if other == nil {
		return result
	}

	result.Base = c.Base.extend(other.Base)
	result.Dev = c.Dev.extend(other.Dev)
	result.Prod = c.Prod.extend(other.Prod)

	return result

}

func (c *Config) validate() error {

	if c.Dev == nil && c.Prod == nil {
		return errors.New("Neither dev nor prod configuration was valid")
	}

	if c.Dev != nil {
		if err := c.Dev.validate(true); err != nil {
			return err
		}
	}
	if c.Prod != nil {
		return c.Prod.validate(false)
	}
	return nil
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

	if other.FirebaseProjectId != "" {
		result.FirebaseProjectId = other.FirebaseProjectId
	}

	if other.DisableAdminChecking {
		result.DisableAdminChecking = true
	}

	//Extend adminID, but no duplicates
	adminIdsSet := make(map[string]bool, len(result.AdminUserIds))
	for _, key := range result.AdminUserIds {
		adminIdsSet[key] = true
	}

	for _, key := range other.AdminUserIds {
		if adminIdsSet[key] {
			//Already in the set, don't add a duplicate
			continue
		}
		result.AdminUserIds = append(result.AdminUserIds, key)
	}

	for key, val := range other.StorageConfig {
		result.StorageConfig[key] = val
	}

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

const (
	configFileName       = "config.SECRET.json"
	sampleConfigFileName = "config.SAMPLE.json"
)

func Get() (*Config, error) {

	fileNameToUse := configFileName

	if _, err := os.Stat(configFileName); os.IsNotExist(err) {

		if _, err := os.Stat(sampleConfigFileName); os.IsNotExist(err) {
			return nil, errors.New("Couldn't find a " + configFileName + " in current directory (or a SAMPLE). This file is required. Copy a starter one from boardgame/server/api/config.SAMPLE.json")
		}

		fileNameToUse = sampleConfigFileName

	}

	contents, err := ioutil.ReadFile(fileNameToUse)

	if err != nil {
		return nil, errors.New("Couldn't read config file: " + err.Error())
	}

	var config Config

	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, errors.New("couldn't unmarshal config file: " + err.Error())
	}

	config.derive()

	if err := config.validate(); err != nil {
		return nil, errors.New("Couldn't validate config: " + err.Error())
	}

	return &config, nil

}
