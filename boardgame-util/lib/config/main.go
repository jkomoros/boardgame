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
	"os"
	"path/filepath"
	"strings"
)

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

	return publicConfig.extendWithSecret(privateConfig), nil

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
