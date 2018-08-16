package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

//RawConfig corresponds to the raw input/output from disk without any
//modifications. The derived Config object will use RawConfig's and combine
//them to create the overall Config.
type RawConfig struct {
	Base *ConfigMode
	Dev  *ConfigMode
	Prod *ConfigMode
	//Path is the path this config was loaded up from
	path string
}

//NewRawConfig loads up a raw config given a config.json file on disk.
//Generally you don't use this directly, but instead use Get().
func NewRawConfig(filename string) (*RawConfig, error) {
	if filename == "" {
		return nil, nil
	}

	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, errors.New("Couldn't read config file: " + err.Error())
	}

	var config RawConfig

	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, errors.New("couldn't unmarshal config file: " + err.Error())
	}

	config.path = filename

	return &config, nil
}

//Path returns the filename of the file that this RawConfig represents on
//disk.
func (r *RawConfig) Path() string {
	return r.path
}
