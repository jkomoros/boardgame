package config

import (
	"errors"
	"path/filepath"
	"strconv"
)

//ModeType defines which sub-mode (Base, Dev, or Prod) we're referring to.
type ModeType int

const (
	//TypeBase referrs to the Base sub-mode that underlies both Dev and Prod.
	TypeBase ModeType = iota
	//TypeDev referrs to the Dev sub-mode that builds on top of Base.
	TypeDev
	//TypeProd referrs to the Prod sub-mode that builds on top of Base.
	TypeProd
)

//Config represents the derived config values. It is based on RawConfigs. This
//is the object you generally use directly in your application.
type Config struct {
	Dev             *Mode
	Prod            *Mode
	rawPublicConfig *RawConfig
	rawSecretConfig *RawConfig
	overriders      []OptionOverrider
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

//OptionOverrider is a class of function that is the primary option that
//config.AddOverride takes. They take a configmode and modify some number of
//properties. prodMode is true if it's being called on the prod option, false if
//on dev. A few are defined as conveniences in this package.
type OptionOverrider func(prodMode bool, c *Mode)

//EnableOfflineDevMode returns an OptionOverrider that can be passed to
//config.AddOverride. It simply sets OfflineDevMode to true. It will not
//modify ProdMode, since that value is unsafe in prod mode.
func EnableOfflineDevMode() OptionOverrider {
	//Return a closure mainly just so that EnableOfflineDevMode nests under OptionOverrider in godoc.
	return func(prodMode bool, c *Mode) {
		//OfflineDevMode should never be enabled on prod mode.
		if prodMode {
			return
		}
		c.OfflineDevMode = true
	}
}

//AddOverride takes an OptionOverrider. Overrides are applied to the final Dev
//and Prod configModes after they are derived, passing prodMode of true for
//the prod mode. If you want an override to only apply to dev or prod, the
//OptionOverrider you pass should filter based on prodMode. The changes are
//temporary and will not be persisted to disk when Save() is called. If you
//want changes that will be persisted, see Update.
func (c *Config) AddOverride(o OptionOverrider) {
	c.overriders = append(c.overriders, o)
	c.derive()
}

//Update modifies the config in a specified way. typ and secret select the
//RawConfigMode to modify (creating one if necessary) and then updater applies
//the update. See Set* functions in this package for factories for common
//Updater. After the update, the Config has all of its values rederived to
//reflect the change. If you want changes that will not be persisted, see
//AddOverride.
func (c *Config) Update(typ ModeType, secret bool, updater Updater) error {

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

	c.Prod = prod.Derive(c, true)
	c.Dev = dev.Derive(c, false)

	for _, overrider := range c.overriders {
		overrider(false, c.Dev)
		overrider(true, c.Prod)
	}

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
