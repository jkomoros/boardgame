package main

import (
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

type configRemove struct {
	configModify
}

func configRemoveFactory(base *boardgameUtil, field config.ModeField, fieldType config.ModeFieldType, positional []string) config.Updater {

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

func (c *configRemove) Run(p writ.Path, positional []string) {
	c.configModify.RunWithUpdateFactory(p, positional, configRemoveFactory)
}

func (c *configRemove) Name() string {
	return "remove"
}

func (c *configRemove) Description() string {
	return "Removes the given value to the list for the given field for slice types"
}

func (c *configRemove) Usage() string {
	return strings.Replace(c.configModify.Usage(), "[SUB-KEY] ", "", -1)
}

func (c *configRemove) HelpText() string {

	return c.Name() + " removes the given value from the list of the given field in the current config, for fields that are slice types. No op if that value is not in the list. KEY is not case sensitive.\n\n" +

		"If KEY is of type []string, simply removes the key from the given val if it exists. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringSlice), ",") + "). `Games` is also of this type.\n\n" +

		"See help for the parent command for more information about configuration in general."

}
