package config

import (
	"errors"
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
//produce the result. rawSecret may be nil. In general you don't use this
//directly, but use Get().
func NewConfig(raw, rawSecret *RawConfig) (*Config, error) {
	result := &Config{
		rawPublicConfig: raw,
		rawSecretConfig: rawSecret,
	}
	result.derive()
	if err := result.validate(); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Config) derive() {

	if c.rawPublicConfig == nil {
		return
	}

	base := c.rawPublicConfig.Base
	prod := c.rawPublicConfig.Prod
	dev := c.rawPublicConfig.Dev

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
