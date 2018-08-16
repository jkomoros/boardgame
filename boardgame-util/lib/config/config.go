package config

import (
	"errors"
)

type Config struct {
	Base *ConfigMode
	Dev  *ConfigMode
	Prod *ConfigMode
	//If extendWithPrivate(other) is called, the other's path will be set here
	secretPath string
	//Path is the path this config was loaded up from
	path string
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

//SecretPath returns the path that this config's secret components were
//loaded from, or "" if no secret components.
func (c *Config) SecretPath() string {
	return c.secretPath
}

//extendWithSecret is a wrapper around extend. In addition to normal behavior,
//it sets result's secretPath to be secret's path. This means that the result
//will return the right thing for Path() and SecretPath().
func (c *Config) extendWithSecret(secret *Config) *Config {
	result := c.extend(secret)
	if secret != nil {
		result.secretPath = secret.path
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
