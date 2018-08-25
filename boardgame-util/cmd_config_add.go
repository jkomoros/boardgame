package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigAdd struct {
	ConfigModify
}

func configAddFactory(field config.ConfigModeField, fieldType config.ConfigModeFieldType, positional []string) config.ConfigUpdater {

	switch fieldType {
	case config.FieldTypeStringSlice:
		if len(positional) != 2 {
			errAndQuit("KEY of type []string wants precisely one VAL")
		}
		return config.AddString(field, positional[1])
	case config.FieldTypeGameNode:
		if len(positional) != 2 {
			errAndQuit("games node wants precisely one VAL")
		}
		return config.AddGame(positional[1])
	}

	return nil
}

func (c *ConfigAdd) Run(p writ.Path, positional []string) {
	c.ConfigModify.RunWithUpdateFactory(p, positional, configAddFactory)
}

func (c *ConfigAdd) Name() string {
	return "add"
}

func (c *ConfigAdd) Description() string {
	return "Adds the given value to the list for the given field for slice types"
}

func (c *ConfigAdd) Usage() string {
	return strings.Replace(c.ConfigModify.Usage(), "[SUB-KEY] ", "", -1)
}

func (c *ConfigAdd) HelpText() string {

	return c.Name() + " adds the given value to the list of the given field in the current config, for fields that are slice types. No op if that value is already in the list. KEY is not case sensitive.\n\n" +

		"If KEY is of type []string, simply adds the key to the given val if it doesn't exist. `config set` also has a similar effect for fields of this type. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringSlice), ",") + "). `Games` is also of this type."

}
