/*

config is a simple library that manages config set-up for boardgame-util and
friends, reading from config.json and config.SECRET.json files. See boardgame-
util/README.md for more on the structure of config.json files.

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
	Games                *GameNode
	//GamesList is not intended to be inflated from JSON, but rather is
	//derived based on the contents of Games.
	GamesList []string
}

//derive takes a raw input and creates a struct with fully derived values in
//Dev/Prod ready for use.
func (c *Config) derive() {
	if c.Base == nil {
		return
	}

	c.Prod = c.Base.extend(c.Prod)
	c.Dev = c.Base.extend(c.Dev)

	c.Prod.derive()
	c.Dev.derive()

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

//derive tells the ConfigMode to do its final processing to create any derived
//fields, like GamesList.
func (c *ConfigMode) derive() {

	if c == nil {
		return
	}

	c.GamesList = c.Games.List()

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

	if other.FirebaseProjectId != "" {
		result.FirebaseProjectId = other.FirebaseProjectId
	}

	if other.DisableAdminChecking {
		result.DisableAdminChecking = true
	}

	result.AdminUserIds = mergedStrList(c.AdminUserIds, other.AdminUserIds)

	for key, val := range other.StorageConfig {
		result.StorageConfig[key] = val
	}

	result.Games = result.Games.extend(other.Games)

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
	privateConfigFileName = "config.SECRET.json"
	publicConfigFileName  = "config.PUBLIC.json"
)

func fileNamesToUse() (publicConfig, privateConfig string) {

	if _, err := os.Stat(privateConfigFileName); err == nil {
		privateConfig = privateConfigFileName
	}

	infos, err := ioutil.ReadDir(".")

	if err != nil {
		return "", ""
	}

	foundNames := make(map[string]bool)

	for _, info := range infos {
		if info.Name() == privateConfigFileName {
			continue
		}
		if strings.HasPrefix(info.Name(), "config.") && strings.HasSuffix(info.Name(), ".json") {
			foundNames[info.Name()] = true
		}
	}

	prioritizedNames := []string{
		publicConfigFileName,
		"config.json",
	}

	for _, name := range prioritizedNames {
		if foundNames[name] {
			return name, privateConfig
		}
	}

	//Whatever, return the first one
	for name, _ := range foundNames {
		return name, privateConfig
	}

	//None of the preferred names were found, just return whatever is in
	//publicConfig, privateConfig.
	return

}

func getConfig(filename string) (*Config, error) {

	if filename == "" {
		return nil, nil
	}

	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, errors.New("Couldn't read config file: " + err.Error())
	}

	var config Config

	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, errors.New("couldn't unmarshal config file: " + err.Error())
	}

	return &config, nil
}

func combinedConfig() (*Config, error) {
	publicConfigName, privateConfigName := fileNamesToUse()

	publicConfig, err := getConfig(publicConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get public config: " + err.Error())
	}

	privateConfig, err := getConfig(privateConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get private config: " + err.Error())
	}

	return publicConfig.extend(privateConfig), nil

}

func Get() (*Config, error) {

	config, err := combinedConfig()

	if err != nil {
		return nil, err
	}

	config.derive()

	if err := config.validate(); err != nil {
		return nil, errors.New("Couldn't validate config: " + err.Error())
	}

	return config, nil

}
