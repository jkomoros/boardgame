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
	"path/filepath"
	"strings"
)

type Config struct {
	Base *ConfigMode
	Dev  *ConfigMode
	Prod *ConfigMode
	//If extendWithPrivate(other) is called, the other's path will be set here
	privatePath string
	//Path is the path this config was loaded up from
	path string
}

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

//derive takes a raw input and creates a struct with fully derived values in
//Dev/Prod ready for use.
func (c *Config) derive() {
	if c.Base != nil {
		c.Prod = c.Base.extend(c.Prod)
		c.Dev = c.Base.extend(c.Dev)
	}

	c.Prod.derive(true)
	c.Dev.derive(false)

}

func (c *Config) copy() *Config {

	result := &Config{}

	//Copy over all of the non-deep stuff
	(*result) = *c

	result.Base = c.Base.copy()
	result.Dev = c.Dev.copy()
	result.Prod = c.Prod.copy()

	return result
}

//Path returns the path that this config's public components were loaded from.
func (c *Config) Path() string {
	return c.path
}

//PrivatePath returns the path that this config's private components were
//loaded from, or "" if no private components.
func (c *Config) PrivatePath() string {
	return c.privatePath
}

//extendWithPrivate is a wrapper around extend. In addition to normal
//behavior, it sets result's privatePath to be private's path. This means that
//the result will return the right thing for Path() and PrivatePath().
func (c *Config) extendWithPrivate(private *Config) *Config {
	result := c.extend(private)
	if private != nil {
		result.privatePath = private.path
	}
	return result
}

//extend takes an other config and returns a *new* config where any non-zero
//value for other extends base. If you're extending with private, use
//extendWithPrivate instead.
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

const (
	privateConfigFileName = "config.SECRET.json"
	publicConfigFileName  = "config.PUBLIC.json"
)

func fileNamesToUse(dir string) (publicConfig, privateConfig string, err error) {

	if dir == "" {
		dir = "."
	}

	goPath, err := filepath.Abs(os.Getenv("GOPATH"))

	if err != nil {
		//Gopath isn't set correctly
		return "", "", errors.New("Couldn't absolutize gopath: " + err.Error())
	}

	for {

		abs, err := filepath.Abs(dir)

		if err != nil {
			//Maybe fell off the end of what is a real file?
			return "", "", errors.New("Got err absolutizing search directory: " + dir + " : " + err.Error())
		}

		if !strings.HasPrefix(abs, goPath) {
			return "", "", errors.New("Fell out of gopath without finding config: " + abs)
		}

		public, private := fileNamesToUseInDir(dir)

		if public != "" || private != "" {
			return public, private, nil
		}

		dir = filepath.Join("..", dir)
	}

	return "", "", errors.New("Couldn't find a path")

}

//fileNamesToUseInDir looks for public/private values precisely in the given folder.
func fileNamesToUseInDir(dir string) (publicConfig, privateConfig string) {

	possiblePrivateConfig := filepath.Join(dir, privateConfigFileName)

	if _, err := os.Stat(possiblePrivateConfig); err == nil {
		privateConfig = possiblePrivateConfig
	}

	infos, err := ioutil.ReadDir(dir)

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
			return filepath.Join(dir, name), privateConfig
		}
	}

	//Whatever, return the first one
	for name := range foundNames {
		return filepath.Join(dir, name), privateConfig
	}

	//None of the preferred names were found, just return whatever is in
	//publicConfig, privateConfig. publicConfig is "", privateConfig already
	//has the dir in it
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

	config.path = filename

	return &config, nil
}

func combinedConfig(dir string) (*Config, error) {
	publicConfigName, privateConfigName, err := fileNamesToUse(dir)

	if err != nil {
		return nil, errors.New("Couldn't get file names to use: " + err.Error())
	}

	publicConfig, err := getConfig(publicConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get public config: " + err.Error())
	}

	privateConfig, err := getConfig(privateConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get private config: " + err.Error())
	}

	return publicConfig.extendWithPrivate(privateConfig), nil

}

//Get fetches a fully realized config by looking for files in dir. If none are
//found, walks upwards in the directory hierarchy (as long as that's still in
//$GOPATH) until it finds a folder that appears to work. If dir is "", working
//directory is assumed.
func Get(dir string) (*Config, error) {

	config, err := combinedConfig(dir)

	if err != nil {
		return nil, err
	}

	config.derive()

	if err := config.validate(); err != nil {
		return nil, errors.New("Couldn't validate config: " + err.Error())
	}

	return config, nil

}
