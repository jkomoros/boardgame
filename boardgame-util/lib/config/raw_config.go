package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

//RawConfig corresponds to the raw input/output from disk without any
//modifications. The derived Config object will use RawConfig's and combine
//them to create the overall Config.
type RawConfig struct {
	Base *RawConfigMode `json:"base,omitempty"`
	Dev  *RawConfigMode `json:"dev,omitempty"`
	Prod *RawConfigMode `json:"prod,omitempty"`
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

	if config.Base != nil {
		config.Base.Games = config.Base.Games.Normalize()
	}

	if config.Dev != nil {
		config.Dev.Games = config.Dev.Games.Normalize()
	}

	if config.Prod != nil {
		config.Prod.Games = config.Prod.Games.Normalize()
	}

	config.path = filename

	return &config, nil
}

//HasContent returns true if there is any content in the RawConfig at all.
func (r *RawConfig) HasContent() bool {
	if r.Base != nil {
		return true
	}
	if r.Dev != nil {
		return true
	}
	if r.Prod != nil {
		return true
	}
	return false
}

//Path returns the filename of the file that this RawConfig represents on
//disk.
func (r *RawConfig) Path() string {
	return r.path
}

//Save saves RawConfig back to disk at Path().
func (r *RawConfig) Save() error {

	if r.Path() == "" {
		return errors.New("No path provided")
	}

	if !r.HasContent() {
		//No content to save. Make sure that nothing exists at that path!
		if _, err := os.Stat(r.Path()); os.IsNotExist(err) {
			//Good, nothing exists there
			return nil
		}
		if err := os.Remove(r.Path()); err != nil {
			return errors.New("No content so tried to remove " + r.Path() + " but got error: " + err.Error())
		}
		return nil
	}

	blob, err := json.MarshalIndent(r, "", "\t")

	if err != nil {
		return errors.New("Couldn't marshal: " + err.Error())
	}

	return ioutil.WriteFile(r.Path(), blob, 0644)

}
