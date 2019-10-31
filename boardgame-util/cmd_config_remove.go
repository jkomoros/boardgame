package main

import (
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

type ConfigRemove struct {
	ConfigModify
}

func configRemoveFactory(base *BoardgameUtil, field config.ModeField, fieldType config.ModeFieldType, positional []string) config.ConfigUpdater {

	switch fieldType {
	case config.FieldTypeStringSlice:
		if len(positional) != 2 {
			base.errAndQuit("KEY of type []string wants precisely one VAL")
		}
		return config.RemoveString(field, positional[1])
	case config.FieldTypeGameNode:
		if len(positional) != 2 {
			base.errAndQuit("games node wants precisely one VAL")
		}
		return config.RemoveGame(positional[1])
	}

	return nil
}

func (c *ConfigRemove) Run(p writ.Path, positional []string) {
	c.ConfigModify.RunWithUpdateFactory(p, positional, configRemoveFactory)
}

func (c *ConfigRemove) Name() string {
	return "remove"
}

func (c *ConfigRemove) Description() string {
	return "Removes the given value to the list for the given field for slice types"
}

func (c *ConfigRemove) Usage() string {
	return strings.Replace(c.ConfigModify.Usage(), "[SUB-KEY] ", "", -1)
}

func (c *ConfigRemove) HelpText() string {

	return c.Name() + " removes the given value from the list of the given field in the current config, for fields that are slice types. No op if that value is not in the list. KEY is not case sensitive.\n\n" +

		"If KEY is of type []string, simply removes the key from the given val if it exists. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringSlice), ",") + "). `Games` is also of this type.\n\n" +

		"See help for the parent command for more information about configuration in general."

}
