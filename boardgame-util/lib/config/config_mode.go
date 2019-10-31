package config

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

//Mode is the final, derived struct holding all of the leaf values in config.
type Mode struct {
	//ConfigMode is primarily just the common config mode values
	ModeCommon
	//Games is not intended to be inflated from JSON, but rather is derived
	//based on the contents of Games. It is OK to use literally as Games in
	//RawConfig, though, because its serialization is a legal GamesNode.
	Games []string `json:"games"`

	parentConfig *Config
}

func (c *Mode) String() string {
	blob, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR, couldn't unmarshal: " + err.Error()
	}
	return string(blob)
}

//GamePackages returns all of the game packages listed in Games[] that are
//valid, with errors for the invalid ones. A wrapper around
//gamepkg.Packages(), that passes the path of the config as optionalBasePath,
//so that relative paths in games listed in config are interpreted as relative
//to the config.json, not whatever working directory boardgame-util is being
//run in.
func (c *Mode) GamePackages() ([]*gamepkg.Pkg, map[string]error) {

	return gamepkg.Packages(c.Games, c.basePath())

}

//AllGamePackages returns either a gamepkg for each listed game, or an error
//if any one of them was invalid. A wrapper around gamepkg.AllPackages(), that
//passes the path of the config as optionalBasePath, so that relative paths in
//games listed in config are interpreted as relative to the config.json, not
//whatever working directory boardgame-util is being run in.
func (c *Mode) AllGamePackages() ([]*gamepkg.Pkg, error) {
	return gamepkg.AllPackages(c.Games, c.basePath())
}

//basePath returns the base path to pass to gamepkg.Packages and friends.
func (c *Mode) basePath() string {
	if c.parentConfig == nil {
		return ""
	}

	path := c.parentConfig.Path()

	if path == "" {
		path = c.parentConfig.SecretPath()
	}

	if path == "" {
		return path
	}

	return filepath.Dir(path)

}

//ParentConfig returns the Config that this ConfigMode is part of.
//Specifically, returns the config that was passed as ParentConfig to
//RawConfigMode.Derive().
func (c *Mode) ParentConfig() *Config {
	return c.parentConfig
}

//OriginAllowed returns whether the given orgiin is allowed, based on the
//configuration of AllowedOrigins, which treats "" and "*" specially.
func (c *Mode) OriginAllowed(origin string) bool {

	originURL, err := url.Parse(origin)

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

		if u.Scheme == originURL.Scheme && u.Host == originURL.Host {
			return true
		}
	}
	return false
}
