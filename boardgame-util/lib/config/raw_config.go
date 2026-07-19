package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
)

// RawConfig corresponds to the raw input/output from disk without any
// modifications. The derived Config object will use RawConfig's and combine
// them to create the overall Config.
type RawConfig struct {
	Base *RawConfigMode `json:"base,omitempty"`
	Dev  *RawConfigMode `json:"dev,omitempty"`
	Prod *RawConfigMode `json:"prod,omitempty"`
	//Path is the path this config was loaded up from
	path string
}

// NewRawConfig loads up a raw config given a config.json file on disk.
// Generally you don't use this directly, but instead use Get(). If create is
// true, then if the file doesn't exist on disk it's not an error, and a blank
// config with that name will be returned.
func NewRawConfig(filename string, create bool) (*RawConfig, error) {
	if filename == "" {
		return nil, nil
	}

	var config RawConfig

	contents, err := os.ReadFile(filename)

	if err != nil {
		// Creating tolerates only an absent file. Permission errors, directories,
		// and other I/O failures must remain loud so they cannot masquerade as an
		// empty configuration that a later Save might overwrite.
		if !create || !os.IsNotExist(err) {
			return nil, fmt.Errorf("couldn't read config file: %w", err)
		}
	} else {
		//If there are file contents, unmarshal
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
	}

	config.path = filename

	return &config, nil
}

// HasContent returns true if there is any content in the RawConfig at all.
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

// Path returns the filename of the file that this RawConfig represents on
// disk.
func (r *RawConfig) Path() string {
	return r.path
}

// Save saves RawConfig back to disk at Path(). If HasContent() returns false
// and Path() doesn't exist yet, no file is saved and a nil error is returned.
func (r *RawConfig) Save() error {
	blob, write, err := r.serializedForSave()
	if err != nil {
		return err
	}
	if !write {
		return nil
	}

	mode := os.FileMode(0o644)
	if filepath.Base(r.Path()) == privateConfigFileName {
		mode = 0o600
	}
	if filepath.Base(r.Path()) == privateConfigFileName {
		if err := fileutil.WriteFileSetAtomic(filepath.Dir(r.Path()), map[string]fileutil.FileSpec{
			filepath.Base(r.Path()): {Contents: blob, Mode: mode, ForceMode: true},
		}, true); err != nil {
			return fmt.Errorf("couldn't save config file: %w", err)
		}
		return nil
	}
	if err := fileutil.WriteFileAtomic(r.Path(), blob, mode); err != nil {
		return fmt.Errorf("couldn't save config file: %w", err)
	}
	return nil
}

func (r *RawConfig) serializedForSave() ([]byte, bool, error) {
	if r.Path() == "" {
		return nil, false, errors.New("no path provided")
	}
	if !r.HasContent() {
		if _, err := os.Stat(r.Path()); os.IsNotExist(err) {
			return nil, false, nil
		} else if err != nil {
			return nil, false, fmt.Errorf("couldn't inspect config file: %w", err)
		}
	}
	blob, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return nil, false, fmt.Errorf("couldn't marshal config: %w", err)
	}
	return blob, true, nil
}
