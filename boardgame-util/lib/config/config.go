package config

import (
	"errors"
	"path/filepath"
	"strconv"
)

type ConfigModeType int

const (
	TypeBase ConfigModeType = iota
	TypeDev
	TypeProd
)

//Config represents the derived config values. It is based on RawConfigs. This
//is the object you generally use directly in your application.
type Config struct {
	Dev             *ConfigMode
	Prod            *ConfigMode
	rawPublicConfig *RawConfig
	rawSecretConfig *RawConfig
}

//NewConfig returns a new, derived config object based on the given raw
//configs, using primarily mode.Extend, mode.Derive in the right order to
//produce the result. Both raw and rawSecret may be nil. In general you don't
//use this directly, but use Get().
func NewConfig(raw, rawSecret *RawConfig) *Config {
	result := &Config{
		rawPublicConfig: raw,
		rawSecretConfig: rawSecret,
	}
	result.derive()
	return result
}

//Update modifies the config in a specified way. typ and secret select the
//RawConfigMode to modify (creating one if necessary) and then updater applies
//the update. See Set* functions in this package for factories for common
//ConfigUpdaters. After the update, the Config has all of its values rederived
//to reflect the change.
func (c *Config) Update(typ ConfigModeType, secret bool, updater ConfigUpdater) error {

	if updater == nil {
		return errors.New("No updater provided. Perhaps the config to the factory was invalid?")
	}

	var rawConfig *RawConfig

	if secret {
		rawConfig = c.rawSecretConfig
		if rawConfig == nil {
			//Need to create one

			path := privateConfigFileName

			if c.rawPublicConfig != nil {
				path = filepath.Join(filepath.Dir(c.Path()), privateConfigFileName)
			}

			rawConfig = &RawConfig{
				path: path,
			}

			c.rawSecretConfig = rawConfig
		}
	} else {
		rawConfig = c.rawPublicConfig

		if rawConfig == nil {
			//Need to create one

			path := publicConfigFileName
			if c.rawSecretConfig != nil {
				path = filepath.Join(filepath.Dir(c.SecretPath()), publicConfigFileName)
			}

			rawConfig = &RawConfig{
				path: path,
			}

			c.rawPublicConfig = rawConfig
		}
	}

	var mode *RawConfigMode

	switch typ {
	case TypeBase:
		mode = rawConfig.Base
		if mode == nil {
			//Create it
			mode = &RawConfigMode{}
			rawConfig.Base = mode
		}
	case TypeDev:
		mode = rawConfig.Dev
		if mode == nil {
			mode = &RawConfigMode{}
			rawConfig.Dev = mode
		}
	case TypeProd:
		mode = rawConfig.Prod
		if mode == nil {
			mode = &RawConfigMode{}
			rawConfig.Prod = mode
		}
	default:
		return errors.New(strconv.Itoa(int(typ)) + " is not a valid type")
	}

	if err := updater(mode, typ); err != nil {
		return errors.New("Updater errored: " + err.Error())
	}

	c.derive()

	return nil

}

func (c *Config) derive() {

	var base *RawConfigMode
	var prod *RawConfigMode
	var dev *RawConfigMode

	if c.rawPublicConfig != nil {
		base = c.rawPublicConfig.Base
		prod = c.rawPublicConfig.Prod
		dev = c.rawPublicConfig.Dev
	}

	if c.rawSecretConfig != nil {
		base = base.Extend(c.rawSecretConfig.Base)
		prod = prod.Extend(c.rawSecretConfig.Prod)
		dev = dev.Extend(c.rawSecretConfig.Dev)
	}

	if base != nil {
		prod = base.Extend(prod)
		dev = base.Extend(dev)
	}

	c.Prod = prod.Derive(true)
	c.Dev = dev.Derive(false)

	return

}

//Save saves the two underlying RawConfigs back to disk. Convenience wrapper
//around RawConfig().Save(), RawSecretConfig.Save()
func (c *Config) Save() error {
	config := c.RawConfig()
	if config != nil {
		if err := config.Save(); err != nil {
			return errors.New("Couldn't save public config: " + err.Error())
		}
	}

	config = c.RawSecretConfig()

	if config != nil {
		if err := config.Save(); err != nil {
			return errors.New("Couldn't save private config: " + err.Error())
		}
	}

	return nil
}

//RawConfig returns the underlying, non-secret config that this derived config
//is based on.
func (c *Config) RawConfig() *RawConfig {
	return c.rawPublicConfig
}

//RawSecretConfig returns the underlying secret config that this derived
//config is based on, or nil if there is no secret component.
func (c *Config) RawSecretConfig() *RawConfig {
	return c.rawSecretConfig
}

//Path returns the path that this config's public components were loaded from.
//Convenience wrapper around c.RawConfig().Path().
func (c *Config) Path() string {
	raw := c.rawPublicConfig
	if raw == nil {
		return ""
	}
	return raw.Path()
}

//SecretPath returns the path that this config's secret components were loaded
//from, or "" if no secret components. Convenience wrapper around
//c.RawSecretConfig().Path().
func (c *Config) SecretPath() string {
	raw := c.rawSecretConfig
	if raw == nil {
		return ""
	}
	return raw.Path()
}
